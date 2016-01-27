// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015 Canonical Ltd
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

package security

import (
	"testing"

	. "gopkg.in/check.v1"

	"github.com/ubuntu-core/snappy/snap"
	"github.com/ubuntu-core/snappy/snap/app"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type SecurityTestSuite struct {
}

var _ = Suite(&SecurityTestSuite{})

func (a *SecurityTestSuite) TestSnappyGetSecurityProfile(c *C) {
	m := snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	b := app.Yaml{Name: "bin/app"}
	ap, err := Profile(&m, b.Name, "/snaps/foo.mvo/1.0/")
	c.Assert(err, IsNil)
	c.Check(ap, Equals, "foo.mvo_bin-app_1.0")
}

func (a *SecurityTestSuite) TestSnappyGetSecurityProfileInvalid(c *C) {
	m := snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	b := app.Yaml{Name: "bin/app"}
	_, err := Profile(&m, b.Name, "/snaps/foo/1.0/")
	c.Assert(err, ErrorMatches, "can not get origin from.*")
}

func (a *SecurityTestSuite) TestSnappyGetSecurityProfileFramework(c *C) {
	m := snap.Info{
		Name:    "foo",
		Version: "1.0",
		Type:    snap.TypeFramework,
	}
	b := app.Yaml{Name: "bin/app"}
	ap, err := Profile(&m, b.Name, "/snaps/foo.mvo/1.0/")
	c.Assert(err, IsNil)
	c.Check(ap, Equals, "foo_bin-app_1.0")
}

func (a *SecurityTestSuite) TestSecurityGenDbusPath(c *C) {
	c.Assert(dbusPath("foo"), Equals, "foo")
	c.Assert(dbusPath("foo bar"), Equals, "foo_20bar")
	c.Assert(dbusPath("foo/bar"), Equals, "foo_2fbar")
}

func (a *SecurityTestSuite) TestSecurityGetAppArmorVars(c *C) {
	appID := &AppID{
		Appname: "foo",
		Version: "1.0",
		AppID:   "id",
		Pkgname: "pkgname",
	}
	c.Assert(appID.AppArmorVars(), Equals, `
# Specified profile variables
@{APP_APPNAME}="foo"
@{APP_ID_DBUS}="id"
@{APP_PKGNAME_DBUS}="pkgname"
@{APP_PKGNAME}="pkgname"
@{APP_VERSION}="1.0"
@{INSTALL_DIR}="{/snaps,/gadget}"
# Deprecated:
@{CLICK_DIR}="{/snaps,/gadget}"`)
}

func (a *SecurityTestSuite) TestSecurityGetAppID(c *C) {
	id, err := NewAppID("pkg_app_1.0")
	c.Assert(err, IsNil)
	c.Assert(id, DeepEquals, &AppID{
		AppID:   "pkg_app_1.0",
		Pkgname: "pkg",
		Appname: "app",
		Version: "1.0",
	})
}

func (a *SecurityTestSuite) TestSecurityGetAppIDInvalid(c *C) {
	_, err := NewAppID("invalid")
	c.Assert(err, Equals, ErrInvalidAppID)
}
