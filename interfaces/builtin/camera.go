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

package builtin

import (
	"fmt"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/udev"
)

const cameraSummary = `allows access to all cameras`

const cameraBaseDeclarationSlots = `
  camera:
    allow-installation:
      slot-snap-type:
        - core
    deny-auto-connection: true
`

const cameraConnectedPlugAppArmor = `
# Until we have proper device assignment, allow access to all cameras
/dev/video[0-9]* rw,

# Allow detection of cameras. Leaks plugged in USB device info
/sys/bus/usb/devices/ r,
/sys/devices/pci**/usb*/**/idVendor r,
/sys/devices/pci**/usb*/**/idProduct r,
/run/udev/data/c81:[0-9]* r, # video4linux (/dev/video*, etc)
`

const cameraConnectedPlugUdev = `
# This file contains udev rules for camera devices.
#
# Do not edit this file, it will be overwritten on updates

KERNEL=="video[0-9]*", TAG+="%s"
`

// The type for camera interface
type cameraInterface struct{}

// Getter for the name of the camera interface
func (iface *cameraInterface) Name() string {
	return "camera"
}

func (iface *cameraInterface) MetaData() interfaces.MetaData {
	return interfaces.MetaData{
		Summary:              cameraSummary,
		ImplicitOnCore:       true,
		ImplicitOnClassic:    true,
		BaseDeclarationSlots: cameraBaseDeclarationSlots,
	}
}

func (iface *cameraInterface) String() string {
	return iface.Name()
}

// Check validity of the defined slot
func (iface *cameraInterface) SanitizeSlot(slot *interfaces.Slot) error {
	// Does it have right type?
	if iface.Name() != slot.Interface {
		panic(fmt.Sprintf("slot is not of interface %q", iface))
	}
	return nil
}

// Checks and possibly modifies a plug
func (iface *cameraInterface) SanitizePlug(plug *interfaces.Plug) error {
	if iface.Name() != plug.Interface {
		panic(fmt.Sprintf("plug is not of interface %q", iface))
	}
	// Currently nothing is checked on the plug side
	return nil
}

func (iface *cameraInterface) AppArmorConnectedPlug(spec *apparmor.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	spec.AddSnippet(cameraConnectedPlugAppArmor)
	return nil
}

func (iface *cameraInterface) UDevConnectedPlug(spec *udev.Specification, plug *interfaces.Plug, plugAttrs map[string]interface{}, slot *interfaces.Slot, slotAttrs map[string]interface{}) error {
	for appName := range plug.Apps {
		tag := udevSnapSecurityName(plug.Snap.Name(), appName)
		spec.AddSnippet(fmt.Sprintf(cameraConnectedPlugUdev, tag))
	}
	return nil
}

func (iface *cameraInterface) AutoConnect(*interfaces.Plug, *interfaces.Slot) bool {
	// Allow what is allowed in the declarations
	return true
}

func init() {
	registerIface(&cameraInterface{})
}
