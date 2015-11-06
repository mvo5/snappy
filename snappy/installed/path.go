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

package installed

import (
	"os"
	"path/filepath"
	"time"

	"github.com/ubuntu-core/snappy/helpers"
)

// InstalledPath represents a installed snap path
type Path string

func (i Path) HasConfig() bool {
	return helpers.FileExists(i.ConfigScript())
}

func (i Path) ConfigScript() string {
	return filepath.Join(string(i), "meta", "hooks", "config")
}

func (i Path) Origin() string {
	ext := filepath.Ext(filepath.Dir(filepath.Clean(string(i))))
	if len(ext) < 2 {
		return ""
	}

	return ext[1:]
}

func (i Path) YamlPath() string {
	return filepath.Join(string(i), "meta", "package.yaml")
}

func (i Path) ReadmePath() string {
	return filepath.Join(string(i), "meta", "readme.md")
}

func (i Path) HashesPath() string {
	return filepath.Join(string(i), "meta", "hashes.yaml")
}

func (i Path) Version() string {
	return filepath.Base(string(i))
}

func (i Path) Size() int64 {
	// FIXME: cache this at install time maybe?
	totalSize := int64(0)
	f := func(_ string, info os.FileInfo, err error) error {
		totalSize += info.Size()
		return err
	}
	filepath.Walk(string(i), f)
	return totalSize
}

func (i Path) Date() time.Time {
	st, err := os.Stat(string(i))
	if err != nil {
		return time.Time{}
	}

	return st.ModTime()
}

func (i Path) RemoveAll() error {
	return os.RemoveAll(string(i))
}

func (i Path) IsActive() (bool, error) {
	allVersionsDir := filepath.Dir(string(i))
	p, err := filepath.EvalSymlinks(filepath.Join(allVersionsDir, "current"))
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	return p == string(i), nil
}
