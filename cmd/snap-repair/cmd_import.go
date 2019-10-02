// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
)

func init() {
	const (
		short = "Import/run a repair assertions from a file"
		long  = ""
	)

	if _, err := parser.AddCommand("import", short, long, &cmdImport{}); err != nil {
		panic(err)
	}

}

type cmdImport struct {
	Positional struct {
		Filename string `positional-arg-name:"<file-name>"`
	} `positional-args:"yes"`
}

func (c *cmdImport) Execute(args []string) error {
	if err := os.MkdirAll(dirs.SnapRunRepairDir, 0755); err != nil {
		return err
	}
	flock, err := osutil.NewFileLock(filepath.Join(dirs.SnapRunRepairDir, "lock"))
	if err != nil {
		return err
	}
	err = flock.TryLock()
	if err == osutil.ErrAlreadyLocked {
		return fmt.Errorf("cannot run, another snap-repair run already executing")
	}
	if err != nil {
		return err
	}
	defer flock.Unlock()

	run := NewRunner()
	err = run.LoadState()
	if err != nil {
		return err
	}
	repair, err := run.Import("canonical", c.Positional.Filename)
	if err != nil {
		return err
	}
	if err := repair.Run(); err != nil {
		return err
	}

	return nil
}
