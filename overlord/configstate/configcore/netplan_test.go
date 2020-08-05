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
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/overlord/configstate/configcore"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/sysconfig"
	"github.com/snapcore/snapd/testutil"
)

type netplanSuite struct {
	configcoreSuite
}

var _ = Suite(&netplanSuite{})

func (s *netplanSuite) TestNetplanHappy(c *C) {
	restore := release.MockOnClassic(false)
	defer restore()

	rootdir := c.MkDir()
	opts := &sysconfig.FilesystemOnlyApplyOptions{Classic: false}

	conf := configcore.PlainCoreConfig(map[string]interface{}{
		"netplan.version": 2,
	})

	err := configcore.FilesystemOnlyApply(rootdir, conf, opts)
	c.Assert(err, IsNil)
	c.Check(filepath.Join(rootdir, "etc/netplan/99-snapd.conf"), testutil.FileEquals, "netplan.version: 2\n")
}
