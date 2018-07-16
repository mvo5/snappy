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

package snap

import (
	"strings"
)

// SnapMapper offers APIs for re-mapping snap names in interfaces and the
// configuration system. The mapper is designed to apply transformations around
// the edges of snapd (state interactions and API interactions) to offer one
// view on the inside of snapd and another view on the outside.
type SnapMapper interface {
	// re-map functions for loading and saving objects in the state.
	IfaceRemapSnapFromState(snapName string) string
	IfaceRemapSnapToState(snapName string) string

	// re-map functions for API requests/responses.
	IfaceRemapSnapFromRequest(snapName string) string
	IfaceRemapSnapToResponse(snapName string) string

	ConfigRemapSnapFromRequest(snapName string) string
	ConfigRemapSnapToResponse(snapName string) string
}

// IdentityMapper implements SnapMapper and performs no transformations at all.
type IdentityMapper struct{}

func (m *IdentityMapper) IfaceRemapSnapFromState(snapName string) string {
	return snapName
}

func (m *IdentityMapper) IfaceRemapSnapToState(snapName string) string {
	return snapName
}

func (m *IdentityMapper) IfaceRemapSnapFromRequest(snapName string) string {
	return snapName
}

func (m *IdentityMapper) IfaceRemapSnapToResponse(snapName string) string {
	return snapName
}

func (m *IdentityMapper) ConfigRemapSnapFromRequest(snapName string) string {
	return snapName
}

func (m *IdentityMapper) ConfigRemapSnapToResponse(snapName string) string {
	return snapName
}

// CoreCoreSystemMapper implements SnapMapper and makes implicit slots
// appear to be on "core" in the state and in memory but as "system" in the API.
//
// NOTE: This mapper can be used to prepare, as an intermediate step, for the
// transition to "snapd" mapper. Using it the state and API layer will look
// exactly the same as with the "snapd" mapper. This can be used to make any
// necessary adjustments the test suite.
type CoreCoreSystemMapper struct {
	IdentityMapper // Embedding the nil mapper allows us to cut on boilerplate.
}

// RemapSnapFromRequest renames the "system" snap to the "core" snap.
//
// This allows us to accept connection and disconnection requests that
// explicitly refer to "core" or using the "system" nickname.
func (m *CoreCoreSystemMapper) IfaceRemapSnapFromRequest(snapName string) string {
	if snapName == "system" {
		return "core"
	}
	return snapName
}

// RemapSnapToResponse renames the "core" snap to the "system" snap.
//
// This allows us to make all the implicitly defined slots, that are really
// associated with the "core" snap to seemingly occupy the "system" snap
// instead.
func (m *CoreCoreSystemMapper) IfaceRemapSnapToResponse(snapName string) string {
	if snapName == "core" {
		return "system"
	}
	return snapName
}

func (m *CoreCoreSystemMapper) ConfigRemapSnapFromRequest(snapName string) string {
	if snapName == "system" {
		return "core"
	}
	return snapName
}

// CoreSnapdSystemMapper implements SnapMapper and makes implicit slots
// appear to be on "core" in the state and on "system" in the API while they
// are on "snapd" in memory.
type CoreSnapdSystemMapper struct {
	IdentityMapper // Embedding the nil mapper allows us to cut on boilerplate.
}

// RemapSnapFromState renames the "core" snap to the "snapd" snap.
//
// This allows modern snapd to load an old state that remembers connections
// between slots on the "core" snap and other snaps. In memory we are actually
// using "snapd" snap for hosting those slots and this lets us stay compatible.
func (m *CoreSnapdSystemMapper) IfaceRemapSnapFromState(snapName string) string {
	if snapName == "core" {
		return "snapd"
	}
	return snapName
}

// RemapSnapToState renames the "snapd" snap to the "core" snap.
//
// This allows the state to stay backwards compatible as all the connections
// seem to refer to the "core" snap, as in pre core{16,18} days where there was
// only one core snap.
func (m *CoreSnapdSystemMapper) IfaceRemapSnapToState(snapName string) string {
	if snapName == "snapd" {
		return "core"
	}
	return snapName
}

func (m *CoreSnapdSystemMapper) ConfigRemapSnapFromRequest(snapName string) string {
	if snapName == "system" {
		return "core"
	}
	return snapName
}

// RemapSnapFromRequest renames the "core" or "system" snaps to the "snapd" snap.
//
// This allows us to accept connection and disconnection requests that
// explicitly refer to "core" or "system" even though we really want them to
// refer to "snapd". Note that this is not fully symmetric with
// RemapSnapToResponse as we explicitly always talk about "system" snap,
// even if the request used "core".
func (m *CoreSnapdSystemMapper) IfaceRemapSnapFromRequest(snapName string) string {
	if snapName == "system" || snapName == "core" {
		return "snapd"
	}
	return snapName
}

// RemapSnapToResponse renames the "snapd" snap to the "system" snap.
//
// This allows us to make all the implicitly defined slots, that are really
// associated with the "snapd" snap to seemingly occupy the "system" snap
// instead. This ties into the concept of using "system" as a nickname (e.g. in
// gadget snap connections).
func (m *CoreSnapdSystemMapper) IfaceRemapSnapToResponse(snapName string) string {
	if snapName == "snapd" {
		return "system"
	}
	return snapName
}

// CaseMapper implements SnapMapper to use upper case internally and lower case externally.
type CaseMapper struct{}

func (m *CaseMapper) IfaceRemapSnapFromState(snapName string) string {
	return strings.ToUpper(snapName)
}

func (m *CaseMapper) IfaceRemapSnapToState(snapName string) string {
	return strings.ToLower(snapName)
}

func (m *CaseMapper) IfaceRemapSnapFromRequest(snapName string) string {
	return strings.ToUpper(snapName)
}

func (m *CaseMapper) IfaceRemapSnapToResponse(snapName string) string {
	return strings.ToLower(snapName)
}

func (m *CaseMapper) ConfigRemapSnapFromRequest(snapName string) string {
	return strings.ToUpper(snapName)
}

func (m *CaseMapper) ConfigRemapSnapToResponse(snapName string) string {
	return strings.ToLower(snapName)
}

// mapper contains the currently active snap mapper.
var mapper SnapMapper = &CoreCoreSystemMapper{}

// MockSnapMapper mocks the currently used snap mapper.
func MockSnapMapper(new SnapMapper) (restore func()) {
	old := mapper
	mapper = new
	return func() { mapper = old }
}

// IfaceRemapSnapFromState renames a snap when loaded from state according to the current mapper.
func IfaceRemapSnapFromState(snapName string) string {
	return mapper.IfaceRemapSnapFromState(snapName)
}

// IfaceRemapSnapToState renames a snap when saving to state according to the current mapper.
func IfaceRemapSnapToState(snapName string) string {
	return mapper.IfaceRemapSnapToState(snapName)
}

// RemapSnapFromRequest  renames a snap as received from an API request according to the current mapper.
func IfaceRemapSnapFromRequest(snapName string) string {
	return mapper.IfaceRemapSnapFromRequest(snapName)
}

// RemapSnapToResponse renames a snap as about to be sent from an API response according to the current mapper.
func IfaceRemapSnapToResponse(snapName string) string {
	return mapper.IfaceRemapSnapToResponse(snapName)
}

// RemapSnapFromRequest  renames a snap as received from an API request according to the current mapper.
func ConfigRemapSnapFromRequest(snapName string) string {
	return mapper.ConfigRemapSnapFromRequest(snapName)
}

// RemapSnapToResponse renames a snap as about to be sent from an API response according to the current mapper.
func ConfigRemapSnapToResponse(snapName string) string {
	return mapper.ConfigRemapSnapToResponse(snapName)
}
