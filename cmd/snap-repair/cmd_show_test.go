// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package main_test

import (
	"os"

	. "gopkg.in/check.v1"

	repair "github.com/snapcore/snapd/cmd/snap-repair"
	"github.com/snapcore/snapd/dirs"
)

func (r *repairSuite) TestShowRepairSingle(c *C) {
	makeMockRepairState(c)

	err := repair.ParseArgs([]string{"show", "canonical-1"})
	c.Check(err, IsNil)
	c.Check(r.Stdout(), Equals, `canonical-1  3  retry
 script:
  #!/bin/sh
  echo retry output
 output:
  repair: canonical-1
  summary: repair one
  
  retry output

`)

}

func (r *repairSuite) TestShowRepairMultiple(c *C) {
	makeMockRepairState(c)

	// repair.ParseArgs() always appends to its internal slice:
	// cmdShow.Positional.Repair. To workaround this we create a
	// new cmdShow here
	err := repair.NewCmdShow("canonical-1", "my-brand-1", "my-brand-2").Execute(nil)
	c.Check(err, IsNil)
	c.Check(r.Stdout(), Equals, `canonical-1  3  retry
 script:
  #!/bin/sh
  echo retry output
 output:
  repair: canonical-1
  summary: repair one
  
  retry output

my-brand-1  1  done
 script:
  #!/bin/sh
  echo done output
 output:
  repair: my-brand-1
  summary: my-brand repair one
  
  done output

my-brand-2  2  skip
 script:
  #!/bin/sh
  echo skip output
 output:
  repair: my-brand-2
  summary: my-brand repair two
  
  skip output

`)
}

func (r *repairSuite) TestShowRepairErrorNoRepairDir(c *C) {
	dirs.SetRootDir(c.MkDir())

	err := repair.NewCmdShow("canonical-1").Execute(nil)
	c.Check(err, ErrorMatches, `cannot find repair "canonical-1"`)
}

func (r *repairSuite) TestShowRepairErrorRepairDirNotReadable(c *C) {
	makeMockRepairState(c)

	err := os.Chmod(dirs.SnapRepairRunDir, 0000)
	c.Assert(err, IsNil)
	defer os.Chmod(dirs.SnapRepairRunDir, 0755)

	err = repair.NewCmdShow("canonical-1").Execute(nil)
	c.Check(err, ErrorMatches, `cannot read snap repair directory: open /.*: permission denied`)
}
