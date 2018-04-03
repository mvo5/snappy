// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package configcore_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/overlord/configstate/configcore"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/testutil"
)

type servicesSuite struct {
	configcoreSuite
	testutil.BaseTest
}

var _ = Suite(&servicesSuite{})

func (s *servicesSuite) SetUpTest(c *C) {
	s.BaseTest.SetUpTest(c)
	s.configcoreSuite.SetUpTest(c)
	dirs.SetRootDir(c.MkDir())
	c.Assert(os.MkdirAll(filepath.Join(dirs.GlobalRootDir, "etc"), 0755), IsNil)
	s.systemctlArgs = nil
	s.BaseTest.AddCleanup(snap.MockSanitizePlugsSlots(func(snapInfo *snap.Info) {}))
}

func (s *servicesSuite) TearDownTest(c *C) {
	dirs.SetRootDir("/")
	s.BaseTest.TearDownTest(c)
}

func (s *servicesSuite) TestConfigureServiceInvalidValue(c *C) {
	err := configcore.SwitchDisableService("ssh", "xxx")
	c.Check(err, ErrorMatches, `option "ssh" has invalid value "xxx"`)
}

func (s *servicesSuite) TestSwitchGenericSystemdServiceNotDisabled(c *C) {
	err := configcore.SwitchGenericSystemdService("rsyslog.service", "false")
	c.Assert(err, IsNil)
	c.Check(s.systemctlArgs, DeepEquals, [][]string{
		{"--root", dirs.GlobalRootDir, "unmask", "rsyslog.service"},
		{"--root", dirs.GlobalRootDir, "enable", "rsyslog.service"},
		{"start", "rsyslog.service"},
	})
}

func (s *servicesSuite) TestSwitchGenericSystemdServiceDisabled(c *C) {
	err := configcore.SwitchGenericSystemdService("rsyslog.service", "true")
	c.Assert(err, IsNil)
	c.Check(s.systemctlArgs, DeepEquals, [][]string{
		{"--root", dirs.GlobalRootDir, "disable", "rsyslog.service"},
		{"--root", dirs.GlobalRootDir, "mask", "rsyslog.service"},
		{"stop", "rsyslog.service"},
		{"show", "--property=ActiveState", "rsyslog.service"},
	})
}

func (s *servicesSuite) TestConfigureServiceDisabledIntegrationSSH(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()

	err := os.MkdirAll(filepath.Join(dirs.GlobalRootDir, "/etc/ssh"), 0755)
	c.Assert(err, IsNil)
	s.systemctlArgs = nil

	err = configcore.Run(&mockConf{
		state: s.state,
		conf: map[string]interface{}{
			"service.ssh.disable": true,
		},
	})
	c.Assert(err, IsNil)
	srv := "ssh.service"
	// SSH is special cased
	sshCanary := filepath.Join(dirs.GlobalRootDir, "/etc/ssh/sshd_not_to_be_run")
	_, err = os.Stat(sshCanary)
	c.Assert(err, IsNil)
	c.Check(s.systemctlArgs, DeepEquals, [][]string{
		{"stop", srv},
		{"show", "--property=ActiveState", srv},
	})
}

func (s *servicesSuite) TestConfigureServiceDisabledIntegrationRsyslog(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()
	s.systemctlArgs = nil

	err := configcore.Run(&mockConf{
		state: s.state,
		conf: map[string]interface{}{
			"service.rsyslog.disable": true,
		},
	})
	c.Assert(err, IsNil)
	srv := "rsyslog.service"
	c.Check(s.systemctlArgs, DeepEquals, [][]string{
		{"--root", dirs.GlobalRootDir, "disable", srv},
		{"--root", dirs.GlobalRootDir, "mask", srv},
		{"stop", srv},
		{"show", "--property=ActiveState", srv},
	})
}

func (s *servicesSuite) TestConfigureServiceDisabledIntegrationConsoleConf(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()

	err := os.MkdirAll(filepath.Join(dirs.GlobalRootDir, "/var/lib/console-conf"), 0755)
	c.Assert(err, IsNil)
	s.systemctlArgs = nil

	err = configcore.Run(&mockConf{
		state: s.state,
		conf: map[string]interface{}{
			"service.console-conf.disable": true,
		},
	})
	c.Assert(err, IsNil)
	canary := filepath.Join(dirs.GlobalRootDir, "/var/lib/console-conf/complete")
	_, err = os.Stat(canary)
	c.Assert(err, IsNil)
	// no service is started/stopped, things must happen at boot time
	c.Check(s.systemctlArgs, IsNil)
}

func (s *servicesSuite) TestConfigureServiceEnableIntegration(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()

	err := os.MkdirAll(filepath.Join(dirs.GlobalRootDir, "/etc/ssh"), 0755)
	c.Assert(err, IsNil)

	for _, service := range []struct {
		cfgName     string
		systemdName string
	}{
		{"ssh", "ssh.service"},
		{"rsyslog", "rsyslog.service"},
	} {
		s.systemctlArgs = nil
		err := configcore.Run(&mockConf{
			state: s.state,
			conf: map[string]interface{}{
				fmt.Sprintf("service.%s.disable", service.cfgName): false,
			},
		})

		c.Assert(err, IsNil)
		srv := service.systemdName
		if service.cfgName == "ssh" {
			// SSH is special cased
			c.Check(s.systemctlArgs, DeepEquals, [][]string{
				{"--root", dirs.GlobalRootDir, "unmask", "sshd.service"},
				{"--root", dirs.GlobalRootDir, "unmask", "ssh.service"},
				{"start", srv},
			})
			sshCanary := filepath.Join(dirs.GlobalRootDir, "/etc/ssh/sshd_not_to_be_run")
			_, err := os.Stat(sshCanary)
			c.Assert(err, ErrorMatches, ".* no such file or directory")
		} else {
			c.Check(s.systemctlArgs, DeepEquals, [][]string{
				{"--root", dirs.GlobalRootDir, "unmask", srv},
				{"--root", dirs.GlobalRootDir, "enable", srv},
				{"start", srv},
			})
		}
	}
}

func (s *servicesSuite) TestConfigureServiceUnsupportedService(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()

	err := configcore.Run(&mockConf{
		state: s.state,
		conf: map[string]interface{}{
			"service.snapd.disable": true,
		},
	})
	c.Assert(err, IsNil)

	// ensure nothing gets enabled/disabled when an unsupported
	// service is set for disable
	c.Check(s.systemctlArgs, IsNil)
}
