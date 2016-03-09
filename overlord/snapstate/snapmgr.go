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
	"strings"

	"github.com/ubuntu-core/snappy/overlord/state"
	"github.com/ubuntu-core/snappy/progress"
	"github.com/ubuntu-core/snappy/snappy"
)

// SnapManager is responsible for the installation and removal of snaps.
type SnapManager struct {
	st *state.State
}

// Manager returns a new snap manager.
func Manager() (*SnapManager, error) {
	return &SnapManager{}, nil
}

// Install initiates a change installing snap.
func (m *SnapManager) Install(snap string) error {
	println("snapmgr installing", snap)
	return nil
}

// Remove initiates a change removing snap.
func (m *SnapManager) Remove(snapSpec string) error {
	m.st.Lock()
	defer m.st.Unlock()

	ch := m.st.NewChange("remove", fmt.Sprintf("Removing %s", snapSpec))
	ch.Set("op", "remove")
	ch.Set("name", snapSpec)

	return nil
}

func doRemove(snapSpec string, meter progress.Meter) error {
	var parts snappy.BySnapVersion

	installed, err := snappy.NewLocalSnapRepository().Installed()
	if err != nil {
		return err
	}
	// Note that "=" is not legal in a snap name or a snap version
	l := strings.Split(snapSpec, "=")
	if len(l) == 2 {
		name := l[0]
		version := l[1]
		parts = snappy.FindSnapsByNameAndVersion(name, version, installed)
	} else {
		parts = snappy.FindSnapsByName(snapSpec, installed)
	}

	if len(parts) == 0 {
		return fmt.Errorf("can not find package: %s", snapSpec)
	}

	legacyoverlord := &snappy.Overlord{}
	for _, part := range parts {
		if err := legacyoverlord.Uninstall(part.(*snappy.SnapPart), meter); err != nil {
			return err
		}
	}

	return nil
}

// Init implements StateManager.Init.
func (m *SnapManager) Init(s *state.State) error {
	println("snapmgr init")
	m.st = s
	return nil
}

// Ensure implements StateManager.Ensure.
func (m *SnapManager) Ensure() error {
	println("snapmgr ensure")
	for _, ch := range m.st.Changes() {
		var op, name string
		ch.Get("op", &op)
		ch.Get("name", &name)
		switch op {
		case "remove":
			doRemove(name, nil)
		default:
			return fmt.Errorf("unsupported operation %s", op)
		}
	}

	return nil
}

// Stop implements StateManager.Stop.
func (m *SnapManager) Stop() error {
	println("snapmgr stop")
	return nil
}
