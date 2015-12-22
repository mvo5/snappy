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

package snapfile

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/ubuntu-core/snappy/pkg/clickdeb"
	"github.com/ubuntu-core/snappy/pkg/info"
	"github.com/ubuntu-core/snappy/pkg/squashfs"
)

// File is the interface to interact with the low-level snap files
type File interface {
	Verify(allowUnauthenticated bool) error
	Close() error

	UnpackWithDropPrivs(targetDir, rootDir string) error
	MetaMember(name string) ([]byte, error)
	ExtractHashes(targetDir string) error
	//Unpack unpacks the src parts to the dst directory
	Unpack(src, dst string) error

	// NeedsMountUnit determines whether it's required to setup
	// a mount unit for the snap when the snap is installed
	NeedsMountUnit() bool

	// Info returns the pkg.Info data
	Info() (info.Info, error)
}

// Open opens a given snap file with the right backend
func Open(path string) (File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open snap: %v", err)
	}
	defer f.Close()

	// look, libmagic!
	header := make([]byte, 20)
	if _, err := f.ReadAt(header, 0); err != nil {
		return nil, fmt.Errorf("cannot read snap: %v", err)
	}
	// Note that we only support little endian squashfs. There
	// is nothing else with squashfs 4.0.
	if bytes.HasPrefix(header, []byte{'h', 's', 'q', 's'}) {
		return squashfs.New(path), nil
	}
	if strings.HasPrefix(string(header), "!<arch>\ndebian") {
		return clickdeb.Open(path)
	}

	return nil, fmt.Errorf("cannot open snap: unknown header: %q", header)
}
