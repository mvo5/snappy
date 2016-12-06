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

package builtin

import (
	"github.com/snapcore/snapd/interfaces"
)

var unity8ContactsPermanentSlotAppArmor = `
# Description: Allow operating as the EDS service. Reserved because this
#  gives privileged access to the system.
# Usage: reserved

# DBus accesses
#include <abstractions/dbus-session-strict>


# Allow binding the service to the requested connection name
dbus (bind)
	bus=session
	name=org.gnome.evolution.dataserver.AddressBook9,
dbus (bind)
	bus=session
	name=org.gnome.evolution.dataserver.Subprocess.Backend.AddressBook*,
dbus (bind)
	bus=session
	name=com.canonical.pim,
`

const unity8ContactsConnectedSlotAppArmor = `
# Allow service to interact with connected clients
# DBus accesses
#include <abstractions/dbus-session-strict>

########################
# EDS - AddressBook
########################
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookFactory,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookFactory
	interface=org.gnome.evolution.dataserver.AddressBookFactory,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookFactory
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookView/**,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookView/**
	interface=org.gnome.evolution.dataserver.AddressBookView,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/AddressBookView/**
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/Subprocess/Backend/AddressBook/**
	interface=org.gnome.evolution.dataserver.Subprocess.Backend,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/Subprocess/Backend/AddressBookView/**
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/Subprocess/**
	interface=org.gnome.evolution.dataserver.AddressBook,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/Subprocess/**
	interface=org.gnome.evolution.dataserver.DirectBook,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/org/gnome/evolution/dataserver/Subprocess/**
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),

########################
# Canonical - AddressBook
########################
dbus (receive, send)
	bus=session
	path=/com/canonical/pim/AddressBook,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/com/canonical/pim/AddressBook
	interface=com.canonical.pim.AddressBook,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
    path=/com/canonical/pim/AddressBook
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/com/canonical/pim/AddressBookView,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
	path=/com/canonical/pim/AddressBookView
	interface=com.canonical.pim.AddressBookView,
	peer=(label=###PLUG_SECURITY_TAGS###),
dbus (receive, send)
	bus=session
    path=/com/canonical/pim/AddressBookView
	interface=org.freedesktop.DBus.*,
	peer=(label=###PLUG_SECURITY_TAGS###),
`

var unity8ContactsConnectedPlugAppArmor = `
# Description: Can access contacts. This policy group is reserved for vetted
#  applications only in this version of the policy. Once LP: #1227821 is
#  fixed, this can be moved out of reserved status.
# Usage: reserved

dbus (receive, send)
     bus=session
     path=/org/gnome/evolution/dataserver/AddressBookFactory
     peer=(label=###SLOT_SECURITY_TAGS###),
dbus (receive, send)
     bus=session
     path=/org/gnome/evolution/dataserver/Subprocess/**
     peer=(label=###SLOT_SECURITY_TAGS###),
dbus (receive, send)
     bus=session
     path=/org/gnome/evolution/dataserver/AddressBookView/**
     peer=(label=###SLOT_SECURITY_TAGS###),

dbus (receive, send)
     bus=session
     path=/com/canonical/pim/AddressBook
     peer=(label=###SLOT_SECURITY_TAGS###),
dbus (receive, send)
     bus=session
     path=/com/canonical/pim/AddressBookView/**
     peer=(label=###SLOT_SECURITY_TAGS###),


# LP: #1319546. Apps shouldn't talk directly to sync-monitor, but allow it for
# now for trusted apps until buteo is integrated with push
# notifications.
dbus (receive, send)
     bus=session
     path=/synchronizer{,/**}
     peer=(label=###SLOT_SECURITY_TAGS###),
`

var unity8ContactsPermanentSlotDBus = `
	<allow own="org.gnome.evolution.dataserver.AddressBook9"/>
	<allow send_destination="org.gnome.evolution.dataserver.AddressBook9"/>
	<allow send_interface="org.gnome.evolution.dataserver.AddressBook"/>
	<allow send_interface="org.gnome.evolution.dataserver.AddressBookView"/>
	<allow send_interface="org.gnome.evolution.dataserver.AddressBookFactory"/>

	<allow own="com.canonical.pim"/>
	<allow send_destination="com.canonical.pim"/>
	<allow send_destination="com.canonical.pim.AddressBook"/>
	<allow send_destination="com.canonical.pim.AddressBookView"/>
`

// NewUnity8ContactsInterface returns a new "unity8-contacts" interface.
func NewUnity8ContactsInterface() interfaces.Interface {
	return &unity8PimCommonInterface{
		name: "unity8-contacts",
		permanentSlotAppArmor: unity8ContactsPermanentSlotAppArmor,
		connectedPlugAppArmor: unity8ContactsConnectedPlugAppArmor,
		connectedSlotAppArmor: unity8ContactsConnectedSlotAppArmor,
		permanentSlotDBus:     unity8ContactsPermanentSlotDBus,
	}
}
