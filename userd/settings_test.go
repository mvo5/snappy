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

package userd_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/testutil"
	"github.com/snapcore/snapd/userd"
)

type settingsSuite struct {
	settings *userd.Settings

	mockXdgSettings *testutil.MockCmd
}

var _ = Suite(&settingsSuite{})

func (s *settingsSuite) SetUpTest(c *C) {
	s.settings = &userd.Settings{}
	s.mockXdgSettings = testutil.MockCommand(c, "xdg-settings", `
if [ "$1" = "get" ] && [ "$2" = "default-web-browser" ];  then
  echo "foo.desktop"
elif [ "$1" = "check" ] && [ "$2" = "default-web-browser" ] && [ "$3" = "foo.desktop" ];  then
  echo yes
elif [ "$1" = "check" ] && [ "$2" = "default-web-browser" ];  then
  echo no
elif [ "$1" = "set" ] && [ "$2" = "default-web-browser" ]; then
  # nothing to do
  exit 0
else
  echo "mock called with unsupported arguments $@"
  exit 1
fi
`)
}

func (s *settingsSuite) TearDownTest(c *C) {
	s.mockXdgSettings.Restore()
}

func (s *settingsSuite) TestGetUnhappy(c *C) {
	for _, t := range []struct {
		setting    string
		errMatcher string
	}{
		{"random-setting", `cannot use setting "random-setting": not allowed`},
		{"invälid", `cannot use setting "invälid": not allowed`},
		{"", `cannot use setting "": not allowed`},
	} {
		_, err := s.settings.Get(t.setting)
		c.Assert(err, ErrorMatches, t.errMatcher)
		c.Assert(s.mockXdgSettings.Calls(), IsNil)
	}
}

func (s *settingsSuite) TestGetHappy(c *C) {
	defaultBrowser, err := s.settings.Get("default-web-browser")
	c.Assert(err, IsNil)
	c.Check(defaultBrowser, Equals, "foo.desktop")
	c.Check(s.mockXdgSettings.Calls(), DeepEquals, [][]string{
		{"xdg-settings", "get", "default-web-browser"},
	})
}

func (s *settingsSuite) TestCheckInvalidSetting(c *C) {
	_, err := s.settings.Check("random-setting", "foo.desktop")
	c.Assert(err, ErrorMatches, `cannot use setting "random-setting": not allowed`)
	c.Assert(s.mockXdgSettings.Calls(), IsNil)
}

func (s *settingsSuite) TestCheckIsDefault(c *C) {
	isDefault, err := s.settings.Check("default-web-browser", "foo.desktop")
	c.Assert(err, IsNil)
	c.Check(isDefault, Equals, "yes")
	c.Check(s.mockXdgSettings.Calls(), DeepEquals, [][]string{
		{"xdg-settings", "check", "default-web-browser", "foo.desktop"},
	})
}

func (s *settingsSuite) TestCheckNoDefault(c *C) {
	isDefault, err := s.settings.Check("default-web-browser", "bar.desktop")
	c.Assert(err, IsNil)
	c.Check(isDefault, Equals, "no")
	c.Check(s.mockXdgSettings.Calls(), DeepEquals, [][]string{
		{"xdg-settings", "check", "default-web-browser", "bar.desktop"},
	})
}

func (s *settingsSuite) TestSetInvalidSetting(c *C) {
	err := s.settings.Set("random-setting", "foo.desktop")
	c.Assert(err, ErrorMatches, `cannot use setting "random-setting": not allowed`)
	c.Assert(s.mockXdgSettings.Calls(), IsNil)
}

func (s *settingsSuite) TestSetUserDeclined(c *C) {
	mockZenity := testutil.MockCommand(c, "zenity", "false")
	defer mockZenity.Restore()

	err := s.settings.Set("default-web-browser", "bar.desktop")
	c.Assert(err, ErrorMatches, `cannot set setting: user declined`)
	c.Check(s.mockXdgSettings.Calls(), IsNil)
	c.Check(mockZenity.Calls(), DeepEquals, [][]string{
		{"zenity", "--question", "--text=Allow changing setting \"default-web-browser\" to \"bar.desktop\" ?"},
	})

}

func (s *settingsSuite) TestSetUserAccepts(c *C) {
	mockZenity := testutil.MockCommand(c, "zenity", "true")
	defer mockZenity.Restore()

	err := s.settings.Set("default-web-browser", "foo.desktop")
	c.Assert(err, IsNil)
	c.Check(s.mockXdgSettings.Calls(), DeepEquals, [][]string{
		{"xdg-settings", "set", "default-web-browser", "foo.desktop"},
	})
	c.Check(mockZenity.Calls(), DeepEquals, [][]string{
		{"zenity", "--question", "--text=Allow changing setting \"default-web-browser\" to \"foo.desktop\" ?"},
	})

}
