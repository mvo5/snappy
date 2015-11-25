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

package release_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/ubuntu-core/snappy/release"
	"github.com/ubuntu-core/snappy/testutil"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type ReleaseTestSuite struct {
	testutil.BaseTest
}

var _ = Suite(&ReleaseTestSuite{})

func (s *ReleaseTestSuite) TestSetup(c *C) {
	c.Assert(release.Setup(c.MkDir()), IsNil)
	c.Check(release.String(), Equals, "rolling-core")
	rel := release.Get()
	c.Check(rel.Flavor, Equals, "core")
	c.Check(rel.Series, Equals, "rolling")
	c.Check(rel.Channel, Equals, "edge")
}

func (s *ReleaseTestSuite) TestOverride(c *C) {
	rel := release.Release{Flavor: "personal", Series: "10.06", Channel: "beta"}
	release.Override(rel)
	c.Check(release.String(), Equals, "10.06-personal")
	c.Check(release.Get(), DeepEquals, rel)
}

func (a *ReleaseTestSuite) makeMockLsbRelease(c *C) {
	mockLsbRelease := filepath.Join(c.MkDir(), "mock-lsb-release")
	s := `
DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=18.09
DISTRIB_CODENAME=awsome
DISTRIB_DESCRIPTION=I'm not real!
`
	err := ioutil.WriteFile(mockLsbRelease, []byte(s), 0644)
	c.Assert(err, IsNil)

	// override lsb-release
	realLsbRelease := release.LsbReleasePath()
	a.AddCleanup(func() { release.SetLsbReleasePath(realLsbRelease) })

	release.SetLsbReleasePath(mockLsbRelease)
}

func (a *ReleaseTestSuite) TestReadLsb(c *C) {
	a.makeMockLsbRelease(c)

	lsb, err := release.ReadLsb()
	c.Assert(err, IsNil)
	c.Assert(lsb.ID, Equals, "Ubuntu")
	c.Assert(lsb.Release, Equals, "18.09")
	c.Assert(lsb.Codename, Equals, "awsome")
}

func (a *ReleaseTestSuite) TestReadLsbNotFound(c *C) {
	a.makeMockLsbRelease(c)
	os.Remove(release.LsbReleasePath())

	_, err := release.ReadLsb()
	c.Assert(err, ErrorMatches, "can not read lsb-release:.*")
}
