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

package abstract

import (
	"time"

	"github.com/ubuntu-core/snappy/pkg"
	"github.com/ubuntu-core/snappy/progress"
)

// Part representation of a snappy part
type Part interface {

	// query
	Name() string
	Version() string
	Description() string
	Origin() string
	Vendor() string

	Hash() string
	IsActive() bool
	IsInstalled() bool
	// Will become active on the next reboot
	NeedsReboot() bool

	// returns the date when the snap was last updated
	Date() time.Time

	// returns the channel of the part
	Channel() string

	// returns the path to the icon (local or uri)
	Icon() string

	// Returns app, framework, core
	Type() pkg.Type

	InstalledSize() int64
	DownloadSize() int64

	// Install the snap
	Install(pb progress.Meter, flags pkg.InstallFlags) (name string, err error)
	// Uninstall the snap
	Uninstall(pb progress.Meter) error
	// Config takes a yaml configuration and returns the full snap
	// config with the changes. Note that "configuration" may be empty.
	Config(configuration []byte) (newConfig string, err error)
	// make an inactive part active, or viceversa
	SetActive(bool, progress.Meter) error

	// get the list of frameworks needed by the part
	Frameworks() ([]string, error)
}
