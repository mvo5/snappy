// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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
	"github.com/jessevdk/go-flags"

	"launchpad.net/snappy/dirs"
	"launchpad.net/snappy/logger"
	"launchpad.net/snappy/priv"
)

// withMutex runs the given function with a filelock mutex
func withMutex(f func() error) error {
	return priv.WithMutex(dirs.SnapLockFile, f)
}

// addOptionDescription will try to find the given longName in the
// options and arguments of the given Command and add a description
//
// if the longName is not found it will panic
func addOptionDescription(arg *flags.Command, longName, description string) {
	for _, opt := range arg.Options() {
		if opt.LongName == longName {
			opt.Description = description
			return
		}
	}
	for _, opt := range arg.Args() {
		if opt.Name == longName {
			opt.Description = description
			return
		}
	}

	logger.Panicf("can not set option description for %#v", longName)
}
