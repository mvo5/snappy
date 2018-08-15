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

package hotplugstate

import (
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/overlord/ifacestate"
	"github.com/snapcore/snapd/overlord/state"
)

// HotplugManager helps running arbitrary commands as tasks.
type HotplugManager struct {
	udevMon  UDevMon
	ifaceMgr *ifacestate.InterfaceManager
}

// Manager returns a new HotplugManager.
func Manager(st *state.State, ifaceMgr *ifacestate.InterfaceManager, runner *state.TaskRunner) *HotplugManager {
	return &HotplugManager{ifaceMgr: ifaceMgr}
}

// Ensure is part of the overlord.StateManager interface.
func (m *HotplugManager) Ensure() error {
	return nil
}

func (m *HotplugManager) Start() error {
	// FIXME: sucky workaround to detect if we are running inside tests
	if dirs.GlobalRootDir != "/" {
		return nil
	}

	udevMon := createUDevMonitor(m.ifaceMgr.HotplugDeviceAdded, m.ifaceMgr.HotplugDeviceRemoved)
	if err := udevMon.Connect(); err != nil {
		return err
	}
	if err := m.udevMon.Run(); err != nil {
		return err
	}
	m.udevMon = udevMon

	return nil
}

func (m *HotplugManager) Stop() error {
	if m.udevMon != nil {
		if err := m.udevMon.Stop(); err != nil {
			return err
		}
		m.udevMon = nil
	}
	return nil
}

var createUDevMonitor = NewUDevMonitor

func MockCreateUDevMonitor() (restore func()) {
	new := func(DeviceAddedCallback, DeviceRemovedCallback) UDevMon { return nil }
	createUDevMonitor = new
	return func() {
		createUDevMonitor = NewUDevMonitor
	}
}
