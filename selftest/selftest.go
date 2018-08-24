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

package selftest

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/logger"
)

var checks = []func() error{
	trySquashfsMount,
}

func writeSelftestFailure() {
	if err := os.MkdirAll(filepath.Dir(selftestPath), 0755); err != nil {
		logger.Noticef("cannot create selftest result dir: %s", err)
	}
	if err := ioutil.WriteFile(dirs.SelftestResult, []byte(err.Error()), 0644); err != nil {
		logger.Noticef("cannot write selftest result: %s", err)
	}
}

func Run() error {
	for _, f := range checks {
		if err := f(); err != nil {
			writeSelftestFailure()
			return err
		}
	}
	if err := os.Remove(dirs.SelftestResult); err != nil && os.IsNotExist(err) {
		logger.Noticef("cannot remove selftest result: %s", err)
	}

	return nil
}
