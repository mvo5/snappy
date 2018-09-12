// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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
	"bytes"
	"fmt"
	"sort"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/interfaces/apparmor"
	"github.com/snapcore/snapd/interfaces/seccomp"
	"github.com/snapcore/snapd/interfaces/udev"
	"github.com/snapcore/snapd/snap"
)

const adbSummary = `allows operating as Android Debug Bridge service`

const adbBaseDeclarationSlots = `
  adb:
    allow-installation:
      slot-snap-type:
        - app
    deny-auto-connection: true
`

var adbVendorList = map[int]string{
	0x03f0: "HP",
	0x03fc: "ECS",
	0x0408: "QUANTA",
	0x0409: "NEC",
	0x0414: "GIGABYTE",
	0x0451: "TI",
	0x0471: "PHILPS",
	0x0482: "KYOCERA",
	0x0489: "FOXCONN",
	0x04b7: "COMPAL",
	0x04c5: "FUJITSU",
	0x04da: "PMC-SIERRA",
	0x04dd: "SHARP",
	0x04e8: "SAMSUNG",
	0x0502: "ACER",
	0x0531: "WACOM",
	0x054c: "SONY",
	0x05c6: "Qualcomm",
	0x067e: "INTERMEC",
	0x091e: "GARMIN-ASUS",
	0x0930: "TOSHIBA",
	0x0955: "NVIDIA",
	0x0b05: "ASUS",
	0x0bb4: "HTC",
	0x0c2e: "HONEYWELL",
	0x0db0: "MSI",
	0x0e79: "ARCHOS",
	0x0e8d: "MTK",
	0x0f1c: "FUNAI",
	0x0fce: "SONY ERICSSON",
	0x1004: "LGE",
	0x109b: "HISENSE",
	0x10a9: "PANTECH",
	0x1219: "COMPALCOMM",
	0x12d1: "HUAWEI",
	0x1662: "POSITIVO",
	0x16d5: "ANYDATA",
	0x17ef: "LENOVO",
	0x18d1: "GOOGLE",
	0x1949: "LAB126",
	0x19a5: "HARRIS",
	0x19d2: "ZTE",
	0x1b8e: "AMLOGIC",
	0x1bbb: "T_AND_A",
	0x1d09: "TECHFAITH",
	0x1d45: "QISDA",
	0x1d4d: "PEGATRON",
	0x1d91: "BYD",
	0x1e85: "GIGASET",
	0x1ebf: "YULONG_COOLPAD",
	0x1f3a: "ALLWINNER",
	0x1f53: "SK TELESYS",
	0x2006: "LENOVOMOBILE",
	0x201e: "HAIER",
	0x2080: "NOOK",
	0x2116: "KT TECH",
	0x2207: "ROCKCHIP",
	0x2237: "KOBO",
	0x2257: "OTGV",
	0x22b8: "MOTOROLA",
	0x22d9: "OPPO",
	0x2314: "INQ_MOBILE",
	0x2340: "TELEEPOCH",
	0x2420: "IRIVER",
	0x24e3: "K-TOUCH",
	0x25e3: "LUMIGON",
	0x2717: "XIAOMI",
	0x271d: "GIONEE",
	0x2836: "OUYA",
	0x2916: "YOTADEVICES",
	0x297f: "EMERGING_TECH",
	0x29a9: "SMARTISAN",
	0x29e4: "PRESTIGIO",
	0x2a45: "MEIZU",
	0x2a47: "BQ",
	0x2a49: "UNOWHY",
	0x2ae5: "FAIRPHONE",
	0x413c: "DELL",
	0x8087: "INTEL", // https://twitter.com/zygoon/status/1032233406564913152
	0xE040: "VIZIO",
}

var adbPermanentSlotAppArmor = `
# Allow adb (server) to use TCP/IP for localhost IPC.
# TODO: once fine-grained network mediation is possible, constrain this to localhost.
network inet stream,
`

var adbConnectedSlotAppArmor = `
# Allow adb (server) to access all usb devices and rely on the device cgroup for mediation.
/dev/bus/usb/[0-9][0-9][0-9]/[0-9][0-9][0-9] rw,

# Allow access to udev meta-data about character devices with major number 189
# as per https://www.kernel.org/doc/Documentation/admin-guide/devices.txt
# those describe "USB serial converters - alternate devices".
/run/udev/data/c189:* r,

# Allow reading the serial number of all the USB devices.
/sys/devices/**/usb*/*/serial r,
`

var adbPermanentSlotSecComp = `
accept
accept4
bind
listen
socket AF_INET SOCK_STREAM -
`

var adbConnectedPlugAppArmor = `
# Allow adb (client) to use TCP/IP for localhost IPC.
# TODO: once fine-grained network mediation is possible, constrain this to localhost.
network inet stream,
`

var adbConnectedPlugSecComp = `
socket AF_INET SOCK_STREAM -
connect
`

type adbInterface struct {
	commonInterface
	sortedVendorIDs []int
}

func (iface *adbInterface) AppArmorPermanentSlot(spec *apparmor.Specification, slot *snap.SlotInfo) error {
	spec.AddSnippet(adbPermanentSlotAppArmor)
	return nil
}

func (iface *adbInterface) AppArmorConnectedSlot(spec *apparmor.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	spec.AddSnippet(adbConnectedSlotAppArmor)
	return nil
}

func (iface *adbInterface) SecCompPermanentSlot(spec *seccomp.Specification, slot *snap.SlotInfo) error {
	spec.AddSnippet(adbPermanentSlotSecComp)
	return nil
}

func (iface *adbInterface) UDevPermanentSlot(spec *udev.Specification, slot *snap.SlotInfo) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Concatenation of all adb udev rules.\n")
	fmt.Fprintf(&buf, "# Do not edit this file, it will be overwritten on update.\n")
	for _, vendorID := range iface.vendorIDs() {
		fmt.Fprintf(&buf, "# %s\n", adbVendorList[vendorID])
		// TODO: consider changing to 0660 once we have support for system groups.
		fmt.Fprintf(&buf, "SUBSYSTEM==\"usb\", ATTR{idVendor}==\"%04x\", MODE=\"0666\"\n", vendorID)
	}
	spec.AddSnippet(buf.String())
	return nil
}

func (iface *adbInterface) UDevConnectedPlug(spec *udev.Specification, plug *interfaces.ConnectedPlug, slot *interfaces.ConnectedSlot) error {
	for _, vendorID := range iface.vendorIDs() {
		spec.TagDevice(fmt.Sprintf("SUBSYSTEM==\"usb\", ATTR{idVendor}==\"%04x\"", vendorID))
	}
	return nil
}

// vendorIDs returns a sorted list of vendor IDs supported by adb interface.
func (iface *adbInterface) vendorIDs() []int {
	if iface.sortedVendorIDs == nil {
		vendorIDs := make([]int, 0, len(adbVendorList))
		for vendorID := range adbVendorList {
			vendorIDs = append(vendorIDs, vendorID)
		}
		sort.Ints(vendorIDs)
		iface.sortedVendorIDs = vendorIDs
	}
	return iface.sortedVendorIDs
}

func init() {
	registerIface(&adbInterface{commonInterface: commonInterface{
		name:                  "adb",
		summary:               adbSummary,
		baseDeclarationSlots:  adbBaseDeclarationSlots,
		connectedPlugAppArmor: adbConnectedPlugAppArmor,
		connectedPlugSecComp:  adbConnectedPlugSecComp,
	}})
}
