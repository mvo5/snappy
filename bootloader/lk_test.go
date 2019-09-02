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

package bootloader_test

import (
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/bootloader"
	"github.com/snapcore/snapd/bootloader/lkenv"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
	"github.com/snapcore/snapd/testutil"
)

type lkImageBuildingTestSuite struct {
	testutil.BaseTest
}

var _ = Suite(&lkImageBuildingTestSuite{})

func (g *lkImageBuildingTestSuite) SetUpTest(c *C) {
	g.BaseTest.SetUpTest(c)
	g.AddCleanup(snap.MockSanitizePlugsSlots(func(snapInfo *snap.Info) {}))

	dirs.SetRootDir(c.MkDir())
	g.AddCleanup(func() { dirs.SetRootDir("") })
}

func (s *lkImageBuildingTestSuite) TestNewLkImageBuildingNolkReturnsNil(c *C) {
	l := bootloader.NewLkImageBuilding()
	c.Assert(l, IsNil)
}

func (s *lkImageBuildingTestSuite) TestNewLkImageBuilding(c *C) {
	bootloader.MockLkImageBuildingFiles(c)
	l := bootloader.NewLkImageBuilding()
	c.Assert(l, NotNil)
}

func (s *lkImageBuildingTestSuite) TestSetGetBootVar(c *C) {
	bootloader.MockLkImageBuildingFiles(c)
	l := bootloader.NewLkImageBuilding()
	bootVars := map[string]string{"snap_mode": "try"}
	l.SetBootVars(bootVars)

	v, err := l.GetBootVars("snap_mode")
	c.Assert(err, IsNil)
	c.Check(v, HasLen, 1)
	c.Check(v["snap_mode"], Equals, "try")
}

func (s *lkImageBuildingTestSuite) TestExtractKernelAssetsUnpacksBootimg(c *C) {
	bootloader.MockLkImageBuildingFiles(c)
	l := bootloader.NewLkImageBuilding()

	c.Assert(l, NotNil)

	files := [][]string{
		{"kernel.img", "I'm a kernel"},
		{"initrd.img", "...and I'm an initrd"},
		{"boot.img", "...and I'm an boot image"},
		{"dtbs/foo.dtb", "g'day, I'm foo.dtb"},
		{"dtbs/bar.dtb", "hello, I'm bar.dtb"},
		// must be last
		{"meta/kernel.yaml", "version: 4.2"},
	}
	si := &snap.SideInfo{
		RealName: "ubuntu-kernel",
		Revision: snap.R(42),
	}
	fn := snaptest.MakeTestSnapWithFiles(c, packageKernel, files)
	snapf, err := snap.Open(fn)
	c.Assert(err, IsNil)

	info, err := snap.ReadInfoFromSnapFile(snapf, si)
	c.Assert(err, IsNil)

	err = l.ExtractKernelAssets(info, snapf)
	c.Assert(err, IsNil)

	// kernel is *not* here
	bootimg := filepath.Join(dirs.GlobalRootDir, "boot", "lk", "boot.img")
	c.Assert(osutil.FileExists(bootimg), Equals, true)
}

func (s *lkImageBuildingTestSuite) TestExtractKernelAssetsUnpacksCustomBootimg(c *C) {
	bootloader.MockLkImageBuildingFiles(c)
	l := bootloader.NewLkImageBuilding()

	c.Assert(l, NotNil)

	// first configure custom boot image file name
	env := lkenv.NewEnv(l.ConfigFile())
	env.Load()
	env.ConfigureBootimgName("boot-2.img")
	err := env.Save()
	c.Assert(err, IsNil)

	files := [][]string{
		{"kernel.img", "I'm a kernel"},
		{"initrd.img", "...and I'm an initrd"},
		{"boot-2.img", "...and I'm an boot image"},
		{"dtbs/foo.dtb", "g'day, I'm foo.dtb"},
		{"dtbs/bar.dtb", "hello, I'm bar.dtb"},
		// must be last
		{"meta/kernel.yaml", "version: 4.2"},
	}
	si := &snap.SideInfo{
		RealName: "ubuntu-kernel",
		Revision: snap.R(42),
	}
	fn := snaptest.MakeTestSnapWithFiles(c, packageKernel, files)
	snapf, err := snap.Open(fn)
	c.Assert(err, IsNil)

	info, err := snap.ReadInfoFromSnapFile(snapf, si)
	c.Assert(err, IsNil)

	err = l.ExtractKernelAssets(info, snapf)
	c.Assert(err, IsNil)

	// kernel is *not* here
	bootimg := filepath.Join(dirs.GlobalRootDir, "boot", "lk", "boot-2.img")
	c.Assert(osutil.FileExists(bootimg), Equals, true)
}
