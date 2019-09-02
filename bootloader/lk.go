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

// lkImageBuilding deals with the little-kernel bootloader during
// image-building (`snap prepare-image` specifically). Here lk
// will write the bootsel files. In a real LK system it will write
// files to partitions.
type lkImageBuilding struct{}

func newLkImageBuilding() Bootloader {
	e := &lkImageBuilding{}
	if !osutil.FileExists(e.envFile()) {
		return nil
	}
	return e
}

func (l *lkImageBuilding) Name() string {
	return "lk"
}

func (l *lkImageBuilding) dir() string {
	// XXX: this check is ugly - but needed because we don't know
	//      if we are in the prepare image phase otherwise
	// XXX2: we could instead check if we have a "bootsel" partition
	//       but that would break "snap prepare-image" on a ubuntu-core
	//       LK machine
	if dirs.GlobalRootDir != "/" {
		return filepath.Join(dirs.GlobalRootDir, "/boot/lk/")
	}
	return ""
}

func (l *lkImageBuilding) SetBootVars(values map[string]string) error {
	return lkSetBootVars(l.envFile(), values)
}

func (l *lkImageBuilding) GetBootVars(names ...string) (map[string]string, error) {
	return lkGetBootVars(l.envFile(), names...)
}

func (l *lkImageBuilding) PrepareImage(gadgetDir string, bootVars map[string]string) error {
	blName := l.Name()
	blConfigFile := l.envFile()
	if err := simplePrepareImage(blName, blConfigFile, gadgetDir); err != nil {
		return err
	}
	return l.SetBootVars(bootVars)
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

type lk struct{}

// newLk create a new lk bootloader object
func newLk() Bootloader {
	e := &lk{}
	if !osutil.FileExists(e.envFile()) {
		return nil
	}
	return e
}

func (l *lk) Name() string {
	return "lk"
}

func (l *lk) SetBootVars(values map[string]string) error {
	return lkSetBootVars(l.envFile(), values)
}

func (l *lk) GetBootVars(names ...string) (map[string]string, error) {
	return lkGetBootVars(l.envFile(), names...)
}

func (l *lk) PrepareImage(string, map[string]string) error {
	return errPrepareImageNothingToDo
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

func lkGetBootVars(envFile string, names ...string) (map[string]string, error) {
	out := make(map[string]string)

	env := lkenv.NewEnv(envFile)
	if err := env.Load(); err != nil {
		return nil, err
	}

	for _, name := range names {
		out[name] = env.Get(name)
	}

	return out, nil
}

func lkSetBootVars(envFile string, values map[string]string) error {
	env := lkenv.NewEnv(envFile)
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
