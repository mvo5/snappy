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

package snap

// ImplicitSlotNames contains the list of slots implicitly added to the OS snap.
var ImplicitSlotNames = []string{
	"firewall-control",
	"home",
	"locale-control",
	"log-observe",
	"mount-observe",
	"network",
	"network-bind",
	"network-control",
	"network-observe",
	"snapd-control",
	"system-observe",
	"timeserver-control",
	"timezone-control",
	"unity7",
	"x11",
}

// AddImplicitSlots adds implicitly defined slots to a given snap.
//
// Only the OS snap has implicit slots.
//
// It is assumed that slots have names matching the interface name. Existing
// slots are not changed, only missing slots are added.
func AddImplicitSlots(snapInfo *Info) {
	if snapInfo.Type != TypeOS {
		return
	}
	for _, ifaceName := range ImplicitSlotNames {
		if _, ok := snapInfo.Slots[ifaceName]; !ok {
			snapInfo.Slots[ifaceName] = &SlotInfo{
				Name:      ifaceName,
				Snap:      snapInfo,
				Interface: ifaceName,
			}
		}
	}
}
