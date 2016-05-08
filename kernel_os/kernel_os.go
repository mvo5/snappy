// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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

package kernel_os

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ubuntu-core/snappy/flags"
	"github.com/ubuntu-core/snappy/osutil"
	"github.com/ubuntu-core/snappy/partition"
	"github.com/ubuntu-core/snappy/progress"
	"github.com/ubuntu-core/snappy/snap"
)

type notifier interface {
	Notify(status string)
}

// override in tests
var findBootloader = partition.FindBootloader

// RemoveKernelAssets removes the unpacked kernel/initrd for the given
// kernel snap
func RemoveKernelAssets(s snap.PlaceInfo, inter notifier) error {
	bootloader, err := findBootloader()
	if err != nil {
		return fmt.Errorf("no not remove kernel assets: %s", err)
	}

	// remove the kernel blob
	blobName := filepath.Base(s.MountFile())
	dstDir := filepath.Join(bootloader.Dir(), blobName)
	if err := os.RemoveAll(dstDir); err != nil {
		return err
	}

	return nil
}

func copyAll(src, dst string) error {
	if output, err := exec.Command("cp", "-a", src, dst).CombinedOutput(); err != nil {
		return fmt.Errorf("cannot copy %q -> %q: %s (%s)", src, dst, err, output)
	}
	return nil
}

// ExtractKernelAssets extracts kernel/initrd/dtb data from the given
// Snap to a versionized bootloader directory so that the bootloader
// can use it.
func ExtractKernelAssets(s *snap.Info, snapf snap.File, flags flags.InstallFlags, inter progress.Meter) error {
	if s.Type != snap.TypeKernel {
		return fmt.Errorf("cannot extract kernel assets from snap type %q", s.Type)
	}

	// sanity check that we have the new kernel format
	_, err := snap.ReadKernelInfo(s)
	if err != nil {
		return err
	}

	bootloader, err := findBootloader()
	if err != nil {
		return fmt.Errorf("cannot extract kernel assets: %s", err)
	}

	if bootloader.Name() == "grub" {
		return nil
	}

	// now do the kernel specific bits
	blobName := filepath.Base(s.MountFile())
	dstDir := filepath.Join(bootloader.Dir(), blobName)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	dir, err := os.Open(dstDir)
	if err != nil {
		return err
	}
	defer dir.Close()

	for _, src := range []string{
		filepath.Join(s.MountDir(), "kernel.img"),
		filepath.Join(s.MountDir(), "initrd.img"),
	} {
		if err := copyAll(src, dstDir); err != nil {
			return err
		}
		if err := dir.Sync(); err != nil {
			return err
		}
	}

	srcDir := filepath.Join(s.MountDir(), "dtbs")
	if osutil.IsDirectory(srcDir) {
		if err := copyAll(srcDir, dstDir); err != nil {
			return err
		}
	}

	return dir.Sync()
}
