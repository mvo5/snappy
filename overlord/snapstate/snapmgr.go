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

// Package snapstate implements the manager and state aspects responsible for the installation and removal of snaps.
package snapstate

import (
	"fmt"

	"github.com/ubuntu-core/snappy/overlord/state"
)

// Install initiates a change installing snap.
func Install(change *state.Change, snap string) error {
	change.State().Lock()
	defer change.State().Unlock()

	// FIXME: make more fine-grained
	tDl := change.NewTask("download-snap", fmt.Sprintf("Installing %q", snap))
	tDl.Set("name", snap)
	tIns := change.NewTask("install-snap", fmt.Sprintf("Installing %q", snap))
	tIns.Set("name", snap)
	// FIMXE: how can tDl communicate the downloaded tmpfile name to tInst
	tIns.WaitFor(tDl)

	return nil
}

// Remove initiates a change removing snap.
func Remove(change *state.Change, snap string) error {
	change.State().Lock()
	defer change.State().Unlock()

	// FIXME: make fine-grained
	t := change.NewTask("remove-snap", fmt.Sprintf("Removing %q", snap))
	t.Set("name", snap)

	return nil
}

// SnapManager is responsible for the installation and removal of snaps.
type SnapManager struct {
	state *state.State

	runner *state.TaskRunner
}

// Manager returns a new snap manager.
func Manager() (*SnapManager, error) {
	return &SnapManager{}, nil
}

func (m *SnapManager) doDownloadSnap(t *state.Task) error {
	var name string
	t.Get("name", &name)
	println("doDownloadSnap", t.Kind(), name)
	return nil
}

func (m *SnapManager) doInstallSnap(t *state.Task) error {
	var name string
	t.Get("name", &name)
	println("doInstallSnap", t.Kind(), name)
	return nil
}

func (m *SnapManager) doRemoveSnap(t *state.Task) error {
	var name string
	t.Get("name", &name)
	println("doRemoveSnap", t.Kind(), name)
	return nil
}

// Init implements StateManager.Init.
func (m *SnapManager) Init(s *state.State) error {
	m.state = s
	m.runner = state.NewTaskRunner(s)

	// FIXME: make more fine grained
	m.runner.AddHandler("download-snap", m.doDownloadSnap)
	m.runner.AddHandler("install-snap", m.doInstallSnap)
	m.runner.AddHandler("remove-snap", m.doRemoveSnap)

	return nil
}

// Ensure implements StateManager.Ensure.
func (m *SnapManager) Ensure() error {
	m.runner.Ensure()
	return nil
}

// Stop implements StateManager.Stop.
func (m *SnapManager) Stop() error {
	m.runner.Stop()
	return nil
}
