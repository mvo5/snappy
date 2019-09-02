// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package bootloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/bootloader/lkenv"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap"
)

type lkBase struct {
	envFilePath string
}

func (l *lkBase) GetBootVars(names ...string) (map[string]string, error) {
	out := make(map[string]string)

	env := lkenv.NewEnv(l.envFilePath)
	if err := env.Load(); err != nil {
		return nil, err
	}

	for _, name := range names {
		out[name] = env.Get(name)
	}

	return out, nil
}

func (l *lkBase) SetBootVars(values map[string]string) error {
	env := lkenv.NewEnv(l.envFilePath)
	if err := env.Load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	// update environment only if something change
	dirty := false
	for k, v := range values {
		// already set to the right value, nothing to do
		if env.Get(k) == v {
			continue
		}
		env.Set(k, v)
		dirty = true
	}

	if dirty {
		return env.Save()
	}

	return nil
}

type lkImageBuilding struct {
	lkBase
}

func newLkImageBuilding() Bootloader {
	e := &lkImageBuilding{}
	if !osutil.FileExists(e.envFile()) {
		return nil
	}
	e.envFilePath = e.envFile()
	return e
}

func (l *lkImageBuilding) Name() string {
	// XXX: use a different name here?
	return "lk"
}

func (l *lkImageBuilding) dir() string {
	if dirs.GlobalRootDir != "/" {
		return filepath.Join(dirs.GlobalRootDir, "/boot/lk/")
	}
	return ""
}

func (l *lkImageBuilding) ConfigFile() string {
	return l.envFile()
}

func (l *lkImageBuilding) envFile() string {
	return filepath.Join(l.dir(), "snapbootsel.bin")
}

// for lk we need to flash boot image to free bootimg partition
// first make sure there is free boot part to use
// if this is image creation, we just extract file
func (l *lkImageBuilding) ExtractKernelAssets(s snap.PlaceInfo, snapf snap.Container) error {
	blobName := filepath.Base(s.MountFile())

	logger.Debugf("ExtractKernelAssets (%s)", blobName)

	env := lkenv.NewEnv(l.envFile())
	if err := env.Load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	bootPartition, err := env.FindFreeBootPartition(blobName)
	if err != nil {
		return err
	}

	// we are preparing image, just extract boot image to bootloader directory
	logger.Debugf("ExtractKernelAssets handling image prepare")
	if err := snapf.Unpack(env.GetBootImageName(), l.dir()); err != nil {
		return fmt.Errorf("Failed to open unpacked %s %v", env.GetBootImageName(), err)
	}

	if err := env.SetBootPartition(bootPartition, blobName); err != nil {
		return err
	}

	return env.Save()
}

func (l *lkImageBuilding) RemoveKernelAssets(s snap.PlaceInfo) error {
	// never needed during image building
	return nil
}

type lk struct {
	lkBase
}

// newLk create a new lk bootloader object
func newLk() Bootloader {
	e := &lk{}
	if !osutil.FileExists(e.envFile()) {
		return nil
	}
	e.envFilePath = e.envFile()
	return e
}

func (l *lk) Name() string {
	return "lk"
}

func (l *lk) ConfigFile() string {
	return l.envFile()
}

func (l *lk) dir() string {
	if dirs.GlobalRootDir != "/" {
		return filepath.Join(dirs.GlobalRootDir, "/dev/disk/by-partlabel/")
	}
	return ""
}

func (l *lk) envFile() string {
	if dirs.GlobalRootDir == "/" {
		// TO-DO: this should be eventually fetched from gadget.yaml
		return filepath.Join(l.dir(), "snapbootsel")
	}
	return ""
}

// for lk we need to flash boot image to free bootimg partition
// first make sure there is free boot part to use
// if this is image creation, we just extract file
func (l *lk) ExtractKernelAssets(s snap.PlaceInfo, snapf snap.Container) error {
	blobName := filepath.Base(s.MountFile())

	logger.Debugf("ExtractKernelAssets (%s)", blobName)

	env := lkenv.NewEnv(l.envFile())
	if err := env.Load(); err != nil && !os.IsNotExist(err) {
		return err
	}

	bootPartition, err := env.FindFreeBootPartition(blobName)
	if err != nil {
		return err
	}

	logger.Debugf("ExtractKernelAssets handling run time usecase")
	// this is live system, extracted bootimg needs to be flashed to
	// free bootimg partition and env has be updated boot slop mapping
	tmpdir, err := ioutil.TempDir("", "bootimg")
	if err != nil {
		return fmt.Errorf("Failed to create tmp directory %v", err)
	}
	defer os.RemoveAll(tmpdir)
	if err := snapf.Unpack(env.GetBootImageName(), tmpdir); err != nil {
		return fmt.Errorf("Failed to unpack %s %v", env.GetBootImageName(), err)
	}
	// write boot.img to free boot partition
	bootimgName := filepath.Join(tmpdir, env.GetBootImageName())
	bif, err := os.Open(bootimgName)
	if err != nil {
		return fmt.Errorf("Failed to open unpacked %s %v", env.GetBootImageName(), err)
	}
	defer bif.Close()
	bpart := filepath.Join(l.dir(), bootPartition)

	bpf, err := os.OpenFile(bpart, os.O_WRONLY, 0660)
	if err != nil {
		return fmt.Errorf("Failed to open boot partition [%s] %v", bpart, err)
	}
	defer bpf.Close()

	buf := make([]byte, 1024)
	for {
		// read by chunks
		n, err := bif.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("Failed to read buffer chunk of %s %v", env.GetBootImageName(), err)
		}
		if n == 0 {
			break
		}
	}

	if err := env.SetBootPartition(bootPartition, blobName); err != nil {
		return err
	}

	return env.Save()
}

func (l *lk) RemoveKernelAssets(s snap.PlaceInfo) error {
	blobName := filepath.Base(s.MountFile())
	logger.Debugf("RemoveKernelAssets (%s)", blobName)
	env := lkenv.NewEnv(l.envFile())
	if err := env.Load(); err != nil && !os.IsNotExist(err) {
		return err
	}
	dirty, _ := env.FreeBootPartition(blobName)
	if dirty {
		return env.Save()
	}
	return nil
}
