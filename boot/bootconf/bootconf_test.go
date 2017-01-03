// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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

package bootconf_test

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/boot/bootconf"
)

func TestBootconf(t *testing.T) { TestingT(t) }

const pi2config = `
# For more options and information see 
# http://www.raspberrypi.org/documentation/configuration/config-txt.md
# Some settings may impact device functionality. See link above for details

kernel=uboot.bin

# uncomment this if your display has a black border of unused pixels visible
# and your display can output without overscan
disable_overscan=1

# uncomment to force a specific HDMI mode (this will force VGA)
hdmi_group=1
#hdmi_mode=1

dtparam=i2c_arm=on
dtparam=i2c_vc=on
#i2s=on
#spi=on
dtparam=act_led_trigger=heartbeat
dtparam=pwr_led_trigger=mmc0

device_tree_address=0x02000000
core_freq=250

`

type piBootConfSuite struct {
	configTxt string
}

var _ = Suite(&piBootConfSuite{})

func (s *piBootConfSuite) SetUpTest(c *C) {
	s.configTxt = filepath.Join(c.MkDir(), "config.txt")
	err := ioutil.WriteFile(s.configTxt, []byte(pi2config), 0644)
	c.Assert(err, IsNil)
}

func (s *piBootConfSuite) TestBootConfGet(c *C) {
	pi2cfg := bootconf.Pi{Path: s.configTxt}
	km, err := pi2cfg.Get([]string{"disable_overscan", "hdmi_group"})
	c.Check(err, IsNil)
	c.Check(km, DeepEquals, map[string]string{
		"disable_overscan": "1",
		"hdmi_group":       "1",
	})
}

func (s *piBootConfSuite) TestBootConfSet(c *C) {
	pi2cfg := bootconf.Pi{Path: s.configTxt}

	err := pi2cfg.Set(map[string]string{
		"disable_overscan": "0",
		"avoid_warnings":   "0",
	})
	c.Check(err, IsNil)

	content, err := ioutil.ReadFile(pi2cfg.Path)
	c.Assert(err, IsNil)
	c.Check(strings.Count(string(content), "\ndisable_overscan=0\n"), Equals, 1)
	c.Check(strings.Count(string(content), "\navoid_warnings=0"), Equals, 1)
}

func (s *piBootConfSuite) TestBootConfSetValidates(c *C) {
	pi2cfg := bootconf.Pi{Path: s.configTxt}
	err := pi2cfg.Set(map[string]string{
		"hdmi_group": "0",
		"wrong-key":  "",
	})
	c.Check(err, ErrorMatches, `cannot use boot config key "wrong-key"`)
}

func (s *piBootConfSuite) TestBootConfGetValidates(c *C) {
	pi2cfg := bootconf.Pi{Path: s.configTxt}
	_, err := pi2cfg.Get([]string{"hdmi_group", "wrong-key"})
	c.Check(err, ErrorMatches, `cannot use boot config key "wrong-key"`)
}
