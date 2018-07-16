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

package snap_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/snap"
)

type mapperSuite struct{}

var _ = Suite(&mapperSuite{})

func (s *mapperSuite) TestIdentityMapper(c *C) {
	restore := snap.MockSnapMapper(&snap.IdentityMapper{})
	defer restore()

	// Nothing is altered.
	c.Assert(snap.IfaceRemapSnapFromState("example"), Equals, "example")
	c.Assert(snap.IfaceRemapSnapToState("example"), Equals, "example")
	c.Assert(snap.IfaceRemapSnapFromRequest("example"), Equals, "example")
	c.Assert(snap.IfaceRemapSnapToResponse("example"), Equals, "example")
}

func (s *mapperSuite) TestCoreCoreSystemMapper(c *C) {
	restore := snap.MockSnapMapper(&snap.CoreCoreSystemMapper{})
	defer restore()

	// Snaps are not renamed when interacting with the state.
	c.Assert(snap.IfaceRemapSnapFromState("core"), Equals, "core")
	c.Assert(snap.IfaceRemapSnapToState("core"), Equals, "core")

	// The "core" snap is renamed to the "system" in API response
	// and back in the API requests.
	c.Assert(snap.IfaceRemapSnapFromRequest("system"), Equals, "core")
	c.Assert(snap.IfaceRemapSnapToResponse("core"), Equals, "system")

	// Other snap names are unchanged.
	c.Assert(snap.IfaceRemapSnapFromState("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapToState("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapFromRequest("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapToResponse("potato"), Equals, "potato")
}

func (s *mapperSuite) TestCoreSnapdSystemMapper(c *C) {
	restore := snap.MockSnapMapper(&snap.CoreSnapdSystemMapper{})
	defer restore()

	// The "snapd" snap is renamed to the "core" in when saving the state
	// and back when loading the state.
	c.Assert(snap.IfaceRemapSnapFromState("core"), Equals, "snapd")
	c.Assert(snap.IfaceRemapSnapToState("snapd"), Equals, "core")

	// The "snapd" snap is renamed to the "system" in API response and back in
	// the API requests.
	c.Assert(snap.IfaceRemapSnapFromRequest("system"), Equals, "snapd")
	c.Assert(snap.IfaceRemapSnapToResponse("snapd"), Equals, "system")

	// The "core" snap is also renamed to "snapd" in API requests, for
	// compatibility.
	c.Assert(snap.IfaceRemapSnapFromRequest("core"), Equals, "snapd")

	// Other snap names are unchanged.
	c.Assert(snap.IfaceRemapSnapFromState("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapToState("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapFromRequest("potato"), Equals, "potato")
	c.Assert(snap.IfaceRemapSnapToResponse("potato"), Equals, "potato")
}

func (s *mapperSuite) TestMappingFunctions(c *C) {
	restore := snap.MockSnapMapper(&snap.CaseMapper{})
	defer restore()

	c.Assert(snap.IfaceRemapSnapFromState("example"), Equals, "EXAMPLE")
	c.Assert(snap.IfaceRemapSnapToState("EXAMPLE"), Equals, "example")
	c.Assert(snap.IfaceRemapSnapFromRequest("example"), Equals, "EXAMPLE")
	c.Assert(snap.IfaceRemapSnapToResponse("EXAMPLE"), Equals, "example")
}
