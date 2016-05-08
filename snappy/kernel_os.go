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

package snappy

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ubuntu-core/snappy/logger"
	"github.com/ubuntu-core/snappy/partition"
	"github.com/ubuntu-core/snappy/snap"
)

// override in tests
var findBootloader = partition.FindBootloader

// setNextBoot will schedule the given os or kernel snap to be used in
// the next boot
func setNextBoot(s *snap.Info) error {
	if s.Type != snap.TypeOS && s.Type != snap.TypeKernel {
		return nil
	}

	bootloader, err := findBootloader()
	if err != nil {
		return fmt.Errorf("cannot set next boot: %s", err)
	}

	var bootvar string
	switch s.Type {
	case snap.TypeOS:
		bootvar = "snappy_os"
	case snap.TypeKernel:
		bootvar = "snappy_kernel"
	}
	blobName := filepath.Base(s.MountFile())
	if err := bootloader.SetBootVar(bootvar, blobName); err != nil {
		return err
	}

	if err := bootloader.SetBootVar("snappy_mode", "try"); err != nil {
		return err
	}

	return nil
}

func kernelOrOsRebootRequired(s *snap.Info) bool {
	if s.Type != snap.TypeKernel && s.Type != snap.TypeOS {
		return false
	}

	bootloader, err := findBootloader()
	if err != nil {
		logger.Noticef("cannot get boot settings: %s", err)
		return false
	}

	var nextBoot, goodBoot string
	switch s.Type {
	case snap.TypeKernel:
		nextBoot = "snappy_kernel"
		goodBoot = "snappy_good_kernel"
	case snap.TypeOS:
		nextBoot = "snappy_os"
		goodBoot = "snappy_good_os"
	}

	nextBootVer, err := bootloader.GetBootVar(nextBoot)
	if err != nil {
		return false
	}
	goodBootVer, err := bootloader.GetBootVar(goodBoot)
	if err != nil {
		return false
	}

	squashfsName := filepath.Base(s.MountFile())
	if nextBootVer == squashfsName && goodBootVer != nextBootVer {
		return true
	}

	return false
}

func nameAndRevnoFromSnap(snap string) (string, int) {
	name := strings.Split(snap, "_")[0]
	revnoNSuffix := strings.Split(snap, "_")[1]
	revno, err := strconv.Atoi(strings.Split(revnoNSuffix, ".snap")[0])
	if err != nil {
		return "", -1
	}
	return name, revno
}

// SyncBoot synchronizes the active kernel and OS snap versions with
// the versions that actually booted. This is needed because a
// system may install "os=v2" but that fails to boot. The bootloader
// fallback logic will revert to "os=v1" but on the filesystem snappy
// still has the "active" version set to "v2" which is
// misleading. This code will check what kernel/os booted and set
// those versions active.
func SyncBoot() error {
	bootloader, err := findBootloader()
	if err != nil {
		return fmt.Errorf("cannot run SyncBoot: %s", err)
	}

	kernelSnap, _ := bootloader.GetBootVar("snappy_kernel")
	osSnap, _ := bootloader.GetBootVar("snappy_os")

	installed, err := (&Overlord{}).Installed()
	if err != nil {
		return fmt.Errorf("cannot run SyncBoot: %s", err)
	}

	overlord := &Overlord{}
	for _, snap := range []string{kernelSnap, osSnap} {
		name, revno := nameAndRevnoFromSnap(snap)
		found := FindSnapsByNameAndRevision(name, revno, installed)
		if len(found) != 1 {
			return fmt.Errorf("cannot SyncBoot, expected 1 snap %q (revno=%d) found %d", snap, revno, len(found))
		}
		if err := overlord.SetActive(found[0], true, nil); err != nil {
			return fmt.Errorf("cannot SyncBoot, cannot make %s active: %s", found[0].Name(), err)
		}
	}

	return nil
}
