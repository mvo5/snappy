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

// Package ifacestate implements the manager and state aspects
// responsible for the maintenance of interfaces the system.
package ifacestate

import (
	"fmt"

	"github.com/ubuntu-core/snappy/interfaces"
	"github.com/ubuntu-core/snappy/interfaces/builtin"
	"github.com/ubuntu-core/snappy/overlord/state"
)

// InterfaceManager is responsible for the maintenance of interfaces in
// the system state.  It maintains interface connections, and also observes
// installed snaps to track the current set of available plugs and slots.
type InterfaceManager struct {
	state  *state.State
	runner *state.TaskRunner
	repo   *interfaces.Repository
}

// Manager returns a new InterfaceManager.
func Manager() (*InterfaceManager, error) {
	return &InterfaceManager{}, nil
}

// Connect initiates a change connecting an interface.
//
func Connect(change *state.Change, plugSnap, plugName, slotSnap, slotName string) error {
	task := change.NewTask("connect", fmt.Sprintf("connect %s:%s to %s:%s",
		plugSnap, plugName, slotSnap, slotName))
	task.Set("slot", interfaces.SlotRef{Snap: slotSnap, Name: slotName})
	task.Set("plug", interfaces.PlugRef{Snap: plugSnap, Name: plugName})
	return nil
}

// Disconnect initiates a change disconnecting an interface.
func (m *InterfaceManager) Disconnect(plugSnap, plugName, slotSnap, slotName string) error {
	return nil
}

// Init implements StateManager.Init.
func (m *InterfaceManager) Init(s *state.State) error {
	repo := interfaces.NewRepository()
	for _, iface := range builtin.Interfaces() {
		if err := repo.AddInterface(iface); err != nil {
			return err
		}
	}
	runner := state.NewTaskRunner(s)
	m.state = s
	m.repo = repo
	m.runner = runner
	m.runner.AddHandler("connect", m.doConnect)
	return nil
}

func (m *InterfaceManager) doConnect(task *state.Task) error {
	task.State().Lock()
	defer task.State().Unlock()

	var slotRef interfaces.SlotRef
	if err := task.Get("slot", &slotRef); err != nil {
		return err
	}
	var plugRef interfaces.PlugRef
	if err := task.Get("plug", &plugRef); err != nil {
		return err
	}
	if err := m.repo.Connect(plugRef.Snap, plugRef.Name, slotRef.Snap, slotRef.Name); err != nil {
		return err
	}
	return nil
}

// Ensure implements StateManager.Ensure.
func (m *InterfaceManager) Ensure() error {
	m.runner.Ensure()
	return nil
}

// Stop implements StateManager.Stop.
func (m *InterfaceManager) Stop() error {
	m.runner.Stop()
	return nil
}
