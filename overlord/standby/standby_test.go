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
package standby_test

import (
	"github.com/snapcore/snapd/overlord/idlestate"
	"github.com/snapcore/snapd/overlord/snapstate"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type standbySuite struct {
	state *state.State
}

var _ = Suite(&standbySuite{})

func (s *standbySuite) SetUpTest(c *C) {
	s.state = state.New(nil)
}
func seeded(st *state.State, b bool) {
	st.Lock()
	st.Set("seeded", b)
	st.Unlock()
}
func (s *standbySuite) TestCanGoSocketActivatedNotSeeded(c *C) {
	seeded(s.state, false)
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, false)
}
func (s *standbySuite) TestCanGoSocketActivatedSeeded(c *C) {
	seeded(s.state, true)
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, true)
}
func (s *standbySuite) TestCanGoSocketActivatedSnaps(c *C) {
	seeded(s.state, true)
	st := s.state
	st.Lock()
	st.Set("seeded", true)
	snapstate.Set(st, "some-snap", &snapstate.SnapState{
		Sequence: []*snap.SideInfo{
			{RealName: "some-snap", Revision: snap.R(1)},
		},
		Current: snap.R(1),
		Active:  true,
	})
	st.Unlock()
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, false)
}
func (s *standbySuite) TestCanGoSocketPendingChanges(c *C) {
	st := s.state
	st.Lock()
	st.Set("seeded", true)
	chg := st.NewChange("foo", "fake change")
	chg.AddTask(st.NewTask("bar", "fake task"))
	c.Assert(chg.Status(), Equals, state.DoStatus)
	st.Unlock()
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, false)
}
func (s *standbySuite) TestCanGoSocketPendingCleans(c *C) {
	st := s.state
	st.Lock()
	st.Set("seeded", true)
	t := st.NewTask("bar", "fake task")
	chg := st.NewChange("foo", "fake change")
	chg.AddTask(t)
	t.SetStatus(state.DoneStatus)
	c.Assert(chg.Status(), Equals, state.DoneStatus)
	c.Assert(t.IsClean(), Equals, false)
	st.Unlock()
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, false)
}
func (s *standbySuite) TestCanGoSocketOnlyDonePendingChanges(c *C) {
	st := s.state
	st.Lock()
	st.Set("seeded", true)
	t := st.NewTask("bar", "fake task")
	chg := st.NewChange("foo", "fake change")
	chg.AddTask(t)
	t.SetStatus(state.DoneStatus)
	t.SetClean()
	c.Assert(chg.Status(), Equals, state.DoneStatus)
	c.Assert(t.IsClean(), Equals, true)
	st.Unlock()
	m := idlestate.Manager(s.state)
	c.Check(m.CanGoSocketActivated(), Equals, true)
}

// FIXME: add test with non-idle connections
