// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
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
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/overlord/configstate/configcore"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/systemd"
	"github.com/snapcore/snapd/testutil"
)

type vitalitySuite struct {
	configcoreSuite

	systemctlCalls  [][]string
	daemonReloadErr error
}

var _ = Suite(&vitalitySuite{})

func (s *vitalitySuite) SetUpTest(c *C) {
	s.configcoreSuite.SetUpTest(c)

	restore := release.MockOnClassic(false)
	s.AddCleanup(restore)

	restore = systemd.MockSystemctl(func(args ...string) ([]byte, error) {
		s.systemctlCalls = append(s.systemctlCalls, args)
		return []byte("ActiveState=inactive"), s.daemonReloadErr
	})
	s.AddCleanup(restore)
}

func (s *vitalitySuite) TestConfigureVitalityUnhappyName(c *C) {
	err := configcore.Run(&mockConf{
		state: s.state,
		changes: map[string]interface{}{
			"resiliance.vitality": "!yf",
		},
	})
	c.Assert(err, ErrorMatches, `cannot set "resiliance.vitality": invalid snap name: "!yf"`)
}

func (s *vitalitySuite) TestConfigureVitality(c *C) {

	err := os.MkdirAll(dirs.SnapServicesDir, 0755)
	c.Assert(err, IsNil)

	// snaps/services that get a vitality score
	toSet := []string{"foo-snap.foo-srv1", "foo-snap.foo-srv2", "bar-snap.bar-srv1"}
	// snaps/services that do *not* get a vitality score
	toNotSet := []string{"other-snap.other-srv"}

	// snaps/services that had a vitality score but no longer have it
	toRemoveSet := []string{"no-longer-set-snap.srv1"}

	// mock existing snaps/services
	for _, srv := range append(append(toSet, toNotSet...), toRemoveSet...) {
		err := ioutil.WriteFile(filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.service", srv)), nil, 0644)
		c.Assert(err, IsNil)
	}
	// mock that some snaps had a vitality config and no longer have it
	for _, srv := range toRemoveSet {
		p := filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.service.d/vitality.conf", srv))
		err = os.MkdirAll(filepath.Dir(p), 0755)
		c.Assert(err, IsNil)
		err = ioutil.WriteFile(p, []byte("[Service]\nOOMScoreAdjust=-888"), 0644)
		c.Assert(err, IsNil)
	}

	// run vitality conf
	err = configcore.Run(&mockConf{
		state: s.state,
		changes: map[string]interface{}{
			"resiliance.vitality": "foo-snap,bar-snap",
		},
	})
	c.Assert(err, IsNil)

	// ensure the affected snaps are set
	for _, name := range toSet {
		p := filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.service.d/vitality.conf", name))
		var expectedScore int
		switch strings.Split(name, ".")[0] {
		case "foo-snap":
			expectedScore = 899
		case "bar-snap":
			expectedScore = 898
		default:
			c.Fatalf("unexpected name %q", name)
		}
		c.Check(p, testutil.FileEquals, fmt.Sprintf("[Service]\nOOMScoreAdjust=-%d\n", expectedScore))
	}

	// and the unaffected snaps do not have a config
	for _, name := range toNotSet {
		p := filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.service.d/vitality.conf", name))
		c.Check(p, testutil.FileAbsent)
	}

	// and the removed vitality settings are honored
	for _, name := range toRemoveSet {
		p := filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.service.d/vitality.conf", name))
		c.Check(p, testutil.FileAbsent)
	}

	// XXX: test that restarts happend etc
	c.Check(s.systemctlCalls[0:1], DeepEquals, [][]string{
		{"daemon-reload"},
	})
}
