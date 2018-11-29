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

package systemd_test

import (
	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/systemd"
	"github.com/snapcore/snapd/testutil"
)

// systemd's testsuite
type runSuite struct{}

var _ = Suite(&runSuite{})

func (s *runSuite) TestRunHappy(c *C) {
	restore := systemd.MockStrutilMakeRandomString(func(int) string {
		return "4" // chosen by a fair dice roll
	})
	defer restore()
	mockSystemdRun := testutil.MockCommand(c, "systemd-run", "")
	defer mockSystemdRun.Restore()
	mockJournalctl := testutil.MockCommand(c, "journalctl", "echo output")
	defer mockJournalctl.Restore()

	output, err := systemd.RunWithOutput("foo", "arg1", "arg2")
	c.Assert(err, IsNil)
	c.Assert(string(output), Equals, "output\n")

	c.Assert(mockSystemdRun.Calls(), DeepEquals, [][]string{
		{"systemd-run", "--wait", "--unit=4.unit", "foo", "arg1", "arg2"},
	})
}
