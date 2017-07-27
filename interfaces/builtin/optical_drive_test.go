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

package builtin_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/builtin"
	"github.com/snapcore/snapd/interfaces/udev"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snaptest"
	"github.com/snapcore/snapd/testutil"
)

type OpticalDriveInterfaceSuite struct {
	iface interfaces.Interface
	slot  *interfaces.Slot
	plug  *interfaces.Plug
}

var _ = Suite(&OpticalDriveInterfaceSuite{
	iface: builtin.MustInterface("optical-drive"),
})

func (s *OpticalDriveInterfaceSuite) SetUpTest(c *C) {
	consumingSnapInfo := snaptest.MockInfo(c, `
name: other
apps:
  app:
    command: foo
    plugs: [optical-drive]
`, nil)
	s.slot = &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap:      &snap.Info{SuggestedName: "core", Type: snap.TypeOS},
			Name:      "optical-drive",
			Interface: "optical-drive",
		},
	}
	s.plug = &interfaces.Plug{PlugInfo: consumingSnapInfo.Plugs["optical-drive"]}
}

func (s *OpticalDriveInterfaceSuite) TestName(c *C) {
	c.Assert(s.iface.Name(), Equals, "optical-drive")
}

func (s *OpticalDriveInterfaceSuite) TestSanitizeSlot(c *C) {
	err := s.iface.SanitizeSlot(s.slot)
	c.Assert(err, IsNil)
	err = s.iface.SanitizeSlot(&interfaces.Slot{SlotInfo: &snap.SlotInfo{
		Snap:      &snap.Info{SuggestedName: "some-snap"},
		Name:      "optical-drive",
		Interface: "optical-drive",
	}})
	c.Assert(err, ErrorMatches, "optical-drive slots are reserved for the operating system snap")
}

func (s *OpticalDriveInterfaceSuite) TestSanitizePlug(c *C) {
	err := s.iface.SanitizePlug(s.plug)
	c.Assert(err, IsNil)
}

func (s *OpticalDriveInterfaceSuite) TestSanitizeIncorrectInterface(c *C) {
	c.Assert(func() { s.iface.SanitizeSlot(&interfaces.Slot{SlotInfo: &snap.SlotInfo{Interface: "app"}}) },
		PanicMatches, `slot is not of interface "optical-drive"`)
	c.Assert(func() { s.iface.SanitizePlug(&interfaces.Plug{PlugInfo: &snap.PlugInfo{Interface: "app"}}) },
		PanicMatches, `plug is not of interface "optical-drive"`)
}

func (s *OpticalDriveInterfaceSuite) TestUsedSecuritySystems(c *C) {
	expectedSnippet1 := `
/dev/sr[0-9]* r,
/dev/scd[0-9]* r,
@{PROC}/sys/dev/cdrom/info r,
`

	expectedSnippet2 := `
# This file contains udev rules for optical drives.
#
# Do not edit this file, it will be overwritten on updates

KERNEL=="sr[0-9]*", NAME="%k", SYMLINK+="scd%n", TAG+="snap_other_app"
`
	// connected plugs have a non-nil security snippet for apparmor
	apparmorSpec := &apparmor.Specification{}
	err := apparmorSpec.AddConnectedPlug(s.iface, s.plug, nil, s.slot, nil)
	c.Assert(err, IsNil)
	c.Assert(apparmorSpec.SecurityTags(), DeepEquals, []string{"snap.other.app"})
	c.Assert(apparmorSpec.SnippetForTag("snap.other.app"), Equals, expectedSnippet1)

	// connected plugs have a non-nil security snippet for udev
	udevSpec := &udev.Specification{}
	c.Assert(udevSpec.AddConnectedPlug(s.iface, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(udevSpec.Snippets(), HasLen, 1)
	snippet := udevSpec.Snippets()[0]
	c.Assert(snippet, Equals, expectedSnippet2)
}

func (s *OpticalDriveInterfaceSuite) TestInterfaces(c *C) {
	c.Check(builtin.Interfaces(), testutil.DeepContains, s.iface)
}
