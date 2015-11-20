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

package partition

import (
	"io/ioutil"
	"os"
	"time"

	. "gopkg.in/check.v1"

	"github.com/mvo5/uboot-go/uenv"
)

const fakeUbootEnvData = `
# This is a snappy variables and boot logic file and is entirely generated and
# managed by Snappy. Modifications may break boot
######
# functions to load kernel, initrd and fdt from various env values
loadfiles=run loadkernel; run loadinitrd; run loadfdt
loadkernel=load mmc ${mmcdev}:${mmcpart} ${loadaddr} ${snappy_ab}/${kernel_file}
loadinitrd=load mmc ${mmcdev}:${mmcpart} ${initrd_addr} ${snappy_ab}/${initrd_file}; setenv initrd_size ${filesize}
loadfdt=load mmc ${mmcdev}:${mmcpart} ${fdtaddr} ${snappy_ab}/dtbs/${fdtfile}

# standard kernel and initrd file names; NB: fdtfile is set early from bootcmd
kernel_file=vmlinuz
initrd_file=initrd.img
fdtfile=am335x-boneblack.dtb

# extra kernel cmdline args, set via mmcroot
snappy_cmdline=init=/lib/systemd/systemd ro panic=-1 fixrtc

# boot logic
# either "a" or "b"; target partition we want to boot
snappy_ab=a
# stamp file indicating a new version is being tried; removed by s-i after boot
snappy_stamp=snappy-stamp.txt
# either "regular" (normal boot) or "try" when trying a new version
snappy_mode=regular
# compat
snappy_trial_boot=0
# if we are trying a new version, check if stamp file is already there to revert
# to other version
snappy_boot=if test "${snappy_mode}" = "try"; then if test -e mmc ${bootpart} ${snappy_stamp}; then if test "${snappy_ab}" = "a"; then setenv snappy_ab "b"; else setenv snappy_ab "a"; fi; else fatwrite mmc ${mmcdev}:${mmcpart} 0x0 ${snappy_stamp} 0; fi; fi; run loadfiles; setenv mmcroot /dev/disk/by-label/system-${snappy_ab} ${snappy_cmdline}; run mmcargs; bootz ${loadaddr} ${initrd_addr}:${initrd_size} ${fdtaddr}
`

func (s *BootloaderTestSuite) makeFakeUbootEnv(c *C) {
	err := os.MkdirAll(bootloaderUbootDir(), 0755)
	c.Assert(err, IsNil)

	// this file just needs to exist
	err = ioutil.WriteFile(bootloaderUbootConfigFile(), []byte(""), 0644)
	c.Assert(err, IsNil)

	// this file needs specific data
	err = ioutil.WriteFile(bootloaderUbootEnvFile(), []byte(fakeUbootEnvData), 0644)
	c.Assert(err, IsNil)
}

func (s *BootloaderTestSuite) TestNewUbootNoUbootReturnsNil(c *C) {
	u := newUboot()
	c.Assert(u, IsNil)
}

func (s *BootloaderTestSuite) TestNewUboot(c *C) {
	s.makeFakeUbootEnv(c)

	u := newUboot()
	c.Assert(u, NotNil)
	c.Assert(u.Name(), Equals, bootloaderNameUboot)
}

func (s *BootloaderTestSuite) TestUbootGetBootVar(c *C) {
	s.makeFakeUbootEnv(c)

	u := newUboot()
	nextBoot, err := u.GetBootVar(bootloaderRootfsVar)
	c.Assert(err, IsNil)
	// the https://developer.ubuntu.com/en/snappy/porting guide says
	// we always use the short names
	c.Assert(nextBoot, Equals, "a")
}

func (s *BootloaderTestSuite) TestUbootGetEnvVar(c *C) {
	s.makeFakeUbootEnv(c)

	u := newUboot()
	c.Assert(u, NotNil)

	v, err := u.GetBootVar(bootloaderBootmodeVar)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "regular")

	v, err = u.GetBootVar(bootloaderRootfsVar)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, "a")
}

func (s *BootloaderTestSuite) TestGetBootloaderWithUboot(c *C) {
	s.makeFakeUbootEnv(c)

	bootloader, err := bootloader()
	c.Assert(err, IsNil)
	c.Assert(bootloader.Name(), Equals, bootloaderNameUboot)
}

func (s *BootloaderTestSuite) TestUbootSetEnvNoUselessWrites(c *C) {
	s.makeFakeUbootEnv(c)

	env, err := uenv.Create(bootloaderUbootFwEnvFile(), 4096)
	c.Assert(err, IsNil)
	env.Set("snappy_ab", "b")
	env.Set("snappy_mode", "regular")
	err = env.Save()
	c.Assert(err, IsNil)

	st, err := os.Stat(bootloaderUbootFwEnvFile())
	c.Assert(err, IsNil)
	time.Sleep(100 * time.Millisecond)

	u := newUboot()
	c.Assert(u, NotNil)

	// note that we set to the same var as above
	err = setBootVarFwEnv(bootloaderRootfsVar, "b")
	c.Assert(err, IsNil)

	env, err = uenv.Open(bootloaderUbootFwEnvFile())
	c.Assert(err, IsNil)
	c.Assert(env.String(), Equals, "snappy_ab=b\nsnappy_mode=regular\n")

	st2, err := os.Stat(bootloaderUbootFwEnvFile())
	c.Assert(err, IsNil)
	c.Assert(st.ModTime(), Equals, st2.ModTime())
}

func (s *BootloaderTestSuite) TestUbootSetBootVarLegacy(c *C) {
	s.makeFakeUbootEnv(c)

	u := newUboot()
	c.Assert(u, NotNil)

	content, err := getBootVarLegacy(bootloaderRootfsVar)
	c.Assert(content, Equals, "a")

	err = setBootVarLegacy(bootloaderRootfsVar, "b")
	c.Assert(err, IsNil)

	content, err = getBootVarLegacy(bootloaderRootfsVar)
	c.Assert(content, Equals, "b")
}

func (s *BootloaderTestSuite) TestUbootSetBootVarFwEnv(c *C) {
	s.makeFakeUbootEnv(c)
	env, err := uenv.Create(bootloaderUbootFwEnvFile(), 4096)
	c.Assert(err, IsNil)
	err = env.Save()
	c.Assert(err, IsNil)

	err = setBootVarFwEnv("key", "value")
	c.Assert(err, IsNil)

	u := newUboot()
	content, err := u.GetBootVar("key")
	c.Assert(err, IsNil)
	c.Assert(content, Equals, "value")
}

func (s *BootloaderTestSuite) TestUbootGetBootVarFwEnv(c *C) {
	s.makeFakeUbootEnv(c)
	env, err := uenv.Create(bootloaderUbootFwEnvFile(), 4096)
	c.Assert(err, IsNil)
	env.Set("key2", "value2")
	err = env.Save()
	c.Assert(err, IsNil)

	u := newUboot()
	content, err := u.GetBootVar("key2")
	c.Assert(err, IsNil)
	c.Assert(content, Equals, "value2")
}
