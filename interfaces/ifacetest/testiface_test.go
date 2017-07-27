// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2017 Canonical Ltd
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

package ifacetest_test

import (
	"fmt"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/dbus"
	"github.com/snapcore/snapd/interfaces/ifacetest"
	"github.com/snapcore/snapd/interfaces/seccomp"
	"github.com/snapcore/snapd/snap"
)

type TestInterfaceSuite struct {
	iface interfaces.Interface
	plug  *interfaces.Plug
	slot  *interfaces.Slot
}

var _ = Suite(&TestInterfaceSuite{
	iface: &ifacetest.TestInterface{InterfaceName: "test"},
	plug: &interfaces.Plug{
		PlugInfo: &snap.PlugInfo{
			Snap:      &snap.Info{SuggestedName: "snap"},
			Name:      "name",
			Interface: "test",
		},
	},
	slot: &interfaces.Slot{
		SlotInfo: &snap.SlotInfo{
			Snap:      &snap.Info{SuggestedName: "snap"},
			Name:      "name",
			Interface: "test",
		},
	},
})

// TestInterface has a working Name() function
func (s *TestInterfaceSuite) TestName(c *C) {
	c.Assert(s.iface.Name(), Equals, "test")
}

// TestInterface has provisions to customize validation
func (s *TestInterfaceSuite) TestValidatePlugError(c *C) {
	iface := &ifacetest.TestInterface{
		InterfaceName: "test",
		ValidatePlugCallback: func(plug *interfaces.Plug, attrs map[string]interface{}) error {
			return fmt.Errorf("validate plug failed")
		},
	}
	err := iface.ValidatePlug(s.plug, nil)
	c.Assert(err, ErrorMatches, "validate plug failed")
}

func (s *TestInterfaceSuite) TestValidateSlotError(c *C) {
	iface := &ifacetest.TestInterface{
		InterfaceName: "test",
		ValidateSlotCallback: func(slot *interfaces.Slot, attrs map[string]interface{}) error {
			return fmt.Errorf("validate slot failed")
		},
	}
	err := iface.ValidateSlot(s.slot, nil)
	c.Assert(err, ErrorMatches, "validate slot failed")
}

// TestInterface doesn't do any sanitization by default
func (s *TestInterfaceSuite) TestSanitizePlugOK(c *C) {
	c.Assert(s.plug.Sanitize(s.iface), IsNil)
}

// TestInterface has provisions to customize sanitization
func (s *TestInterfaceSuite) TestSanitizePlugError(c *C) {
	iface := &ifacetest.TestInterface{
		InterfaceName: "test",
		SanitizePlugCallback: func(plug *interfaces.Plug) error {
			return fmt.Errorf("sanitize plug failed")
		},
	}
	c.Assert(s.plug.Sanitize(iface), ErrorMatches, "sanitize plug failed")
}

// TestInterface doesn't do any sanitization by default
func (s *TestInterfaceSuite) TestSanitizeSlotOK(c *C) {
	c.Assert(s.slot.Sanitize(s.iface), IsNil)
}

// TestInterface has provisions to customize sanitization
func (s *TestInterfaceSuite) TestSanitizeSlotError(c *C) {
	iface := &ifacetest.TestInterface{
		InterfaceName: "test",
		SanitizeSlotCallback: func(slot *interfaces.Slot) error {
			return fmt.Errorf("sanitize slot failed")
		},
	}
	c.Assert(s.slot.Sanitize(iface), ErrorMatches, "sanitize slot failed")
}

// TestInterface hands out empty plug security snippets
func (s *TestInterfaceSuite) TestPlugSnippet(c *C) {
	iface := s.iface.(*ifacetest.TestInterface)

	apparmorSpec := &apparmor.Specification{}
	c.Assert(iface.AppArmorConnectedPlug(apparmorSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(apparmorSpec.Snippets(), HasLen, 0)

	seccompSpec := &seccomp.Specification{}
	c.Assert(iface.SecCompConnectedPlug(seccompSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(seccompSpec.Snippets(), HasLen, 0)

	dbusSpec := &dbus.Specification{}
	c.Assert(iface.DBusConnectedPlug(dbusSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(dbusSpec.Snippets(), HasLen, 0)
}

// TestInterface hands out empty slot security snippets
func (s *TestInterfaceSuite) TestSlotSnippet(c *C) {
	iface := s.iface.(*ifacetest.TestInterface)

	apparmorSpec := &apparmor.Specification{}
	c.Assert(iface.AppArmorConnectedSlot(apparmorSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(apparmorSpec.Snippets(), HasLen, 0)

	seccompSpec := &seccomp.Specification{}
	c.Assert(iface.SecCompConnectedSlot(seccompSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(seccompSpec.Snippets(), HasLen, 0)

	dbusSpec := &dbus.Specification{}
	c.Assert(iface.DBusConnectedSlot(dbusSpec, s.plug, nil, s.slot, nil), IsNil)
	c.Assert(dbusSpec.Snippets(), HasLen, 0)
}

func (s *TestInterfaceSuite) TestAutoConnect(c *C) {
	c.Check(s.iface.AutoConnect(nil, nil), Equals, true)

	iface := &ifacetest.TestInterface{AutoConnectCallback: func(*interfaces.Plug, *interfaces.Slot) bool { return false }}

	c.Check(iface.AutoConnect(nil, nil), Equals, false)
}
