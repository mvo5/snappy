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

/* FIXME:
Feedback

@mvogt This looks more like a Snap than a Path
@mvogt InstalledSnap, maybe
@mvogt For a path, I would expect focus on the path itself, but we (rightfuly) see much richer functionality there

John Lenton
@mvogt @niemeyer if packageYaml were in a package, this would live in that package
John Lenton
@mvogt so I propose you make it that package, even if it doesn't have packageYaml yet

Gustavo Niemeyer
@mvogt Alright, so here is a more complete high-level suggestion
@mvogt The concept hints at something really interesting, but it feels like it isn't really a Path that you are after. For example, IsActive isn't really a property of a path. This looks to me like a first class abstraction for installed snaps as mentioned above.
@mvogt The location for this might be something along the lines of snappy.InstalledSnap, because this is really something that is tied in to the snappy runtime behavior, except today snappy is that catch all package that we are trying to get rid of.
@mvogt I suggest creating a top-level snapsys package whose job is to hold the public snappy-specific functionality that manipulates a local snappy system. In this package we'll hold the sort of higher level abstraction that we were talking about in the context of John's locking work.
@mvogt Once we clean up the snappy package completely, we can move this there.. or perhaps we just keep snapsys, which is not such a bad name and snappy/snapsys ends up better than snappy/snappy..
@mvogt Then, I think we can make that API more high-level over time.. it's fine by me if that's not done right now, though
@mvogt For example, rather than returning the path to the metadata yaml file, why don't we offer a structure out with those details already parsed?
*/

package local

import (
	"os"
	"path/filepath"
	"time"

	"github.com/ubuntu-core/snappy/helpers"
)

// Part represents a installed snap package
type Part string

func (i Part) HasConfig() bool {
	return helpers.FileExists(i.ConfigScript())
}

func (i Part) ConfigScript() string {
	return filepath.Join(string(i), "meta", "hooks", "config")
}

func (i Part) Origin() string {
	ext := filepath.Ext(filepath.Dir(filepath.Clean(string(i))))
	if len(ext) < 2 {
		return ""
	}

	return ext[1:]
}

func (i Part) YamlPath() string {
	return filepath.Join(string(i), "meta", "package.yaml")
}

func (i Part) ReadmePath() string {
	return filepath.Join(string(i), "meta", "readme.md")
}

func (i Part) HashesPath() string {
	return filepath.Join(string(i), "meta", "hashes.yaml")
}

func (i Part) Version() string {
	return filepath.Base(string(i))
}

func (i Part) Size() int64 {
	// FIXME: cache this at install time maybe?
	totalSize := int64(0)
	f := func(_ string, info os.FileInfo, err error) error {
		totalSize += info.Size()
		return err
	}
	filepath.Walk(string(i), f)
	return totalSize
}

func (i Part) Date() time.Time {
	st, err := os.Stat(string(i))
	if err != nil {
		return time.Time{}
	}

	return st.ModTime()
}

func (i Part) RemoveAll() error {
	return os.RemoveAll(string(i))
}

func (i Part) IsActive() (bool, error) {
	allVersionsDir := filepath.Dir(string(i))
	p, err := filepath.EvalSymlinks(filepath.Join(allVersionsDir, "current"))
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}

	return p == string(i), nil
}

