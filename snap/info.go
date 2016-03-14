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

package snap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/osutil"
)

// Info provides information about packages
type Info struct {
	Name    string
	Version string
	Type    Type

	// FIXME: compat with the store, should be "summary"
	Title       string
	Description string

	Channel      string
	IconURI      string
	Date         time.Time
	Origin       string
	Hash         string
	DownloadSize int64

	Revision int64

	URL string
}

func (i *Info) remoteManifestPath() string {
	return filepath.Join(dirs.SnapMetaDir, fmt.Sprintf("%s.%s_%s.manifest", i.Name, i.Origin, i.Version))
}

func (i *Info) SaveManifest() error {
	content, err := yaml.Marshal(i)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dirs.SnapMetaDir, 0755); err != nil {
		return err
	}

	// don't worry about previous contents
	return osutil.AtomicWriteFile(i.remoteManifestPath(), content, 0644, 0)
}
