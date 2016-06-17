// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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

package bind_test

import (
	"os"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/bind"
)

func Test(t *testing.T) {
	TestingT(t)
}

type backendSuite struct {
	backend *bind.Backend
	repo    *interfaces.Repository
	iface   *interfaces.TestInterface
	rootDir string
}

var _ = Suite(&backendSuite{backend: &bind.Backend{}})

func (s *backendSuite) SetUpTest(c *C) {
	s.rootDir = c.MkDir()
	dirs.SetRootDir(s.rootDir)

	err := os.MkdirAll(dirs.SnapBindPolicyDir, 0700)
	c.Assert(err, IsNil)

	s.repo = interfaces.NewRepository()
	s.iface = &interfaces.TestInterface{InterfaceName: "iface"}
	err = s.repo.AddInterface(s.iface)
	c.Assert(err, IsNil)
}

func (s *backendSuite) TearDownTest(c *C) {
	dirs.SetRootDir("/")
}

func (s *backendSuite) TestName(c *C) {
	c.Check(s.backend.Name(), Equals, "bind")
}
