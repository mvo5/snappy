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
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/asserts/snapasserts"
	"github.com/snapcore/snapd/asserts/sysdb"
)

func init() {
	const (
		short = "Verify the given snaps"
		long  = ""
	)

	if _, err := parser.AddCommand("snaps", short, long, &cmdSnaps{}); err != nil {
		panic(err)
	}
}

type cmdSnaps struct {
	Positional struct {
		Assertsdir string   `required:"true"`
		Snaps      []string `required:"true"`
	} `positional-args:"yes"`
}

func readAsserts(fn string, batch *asserts.Batch) ([]*asserts.Ref, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return batch.AddStream(f)
}

func (c *cmdSnaps) Execute(args []string) error {
	path := c.Positional.Assertsdir

	bs := asserts.NewMemoryBackstore()
	cfg := &asserts.DatabaseConfig{
		Backstore: bs,
		Trusted:   sysdb.Trusted(),
	}
	db, err := asserts.OpenDatabase(cfg)
	if err != nil {
		return err
	}

	batch := asserts.NewBatch()
	dc, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}
	for _, fi := range dc {
		fn := filepath.Join(path, fi.Name())
		_, err := readAsserts(fn, batch)
		if err != nil {
			return fmt.Errorf("cannot read assertions: %s", err)
		}
	}
	if err := batch.Commit(db); err != nil {
		return err
	}

	for _, snapFile := range c.Positional.Snaps {
		fmt.Printf("Verifiying %s\n", snapFile)
		if _, err := snapasserts.DeriveSideInfo(snapFile, db); err != nil {
			return err
		}
	}
	return nil
}
