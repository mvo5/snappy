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

package snappy

import (
	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/progress"

	. "gopkg.in/check.v1"
)

type overlordTestSuite struct {
}

var _ = Suite(&overlordTestSuite{})

func (o *overlordTestSuite) SetUpTest(c *C) {
	dirs.SetRootDir(c.MkDir())
}

var helloAppYaml = `name: hello-app
version: 1.0
`

func (o *overlordTestSuite) TestInstalled(c *C) {
	makeInstalledMockSnap(dirs.GlobalRootDir, helloAppYaml)

	overlord := NewOverlord(nil)
	installed, err := overlord.Installed()
	c.Assert(err, IsNil)
	c.Assert(installed, HasLen, 1)
	c.Assert(installed[0].Name(), Equals, "hello-app")
}

func (o *overlordTestSuite) TestNewOverlordWithFunc(c *C) {
	cb := func() progress.Meter {
		return &progress.NullProgress{}
	}

	ol := NewOverlord(cb)
	c.Assert(ol.findMeterCB(), FitsTypeOf, &progress.NullProgress{})
}
