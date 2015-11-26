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

package snappy

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ubuntu-core/snappy/classic"
	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/i18n"
	"github.com/ubuntu-core/snappy/pkg"
	"github.com/ubuntu-core/snappy/progress"
)

// ClassicPart is a fake ubuntu-classic part
type ClassicPart struct {
}

// Type returns pkg.TypeCore for this snap
func (s *ClassicPart) Type() pkg.Type {
	return pkg.TypeApp
}

// Name returns the name
func (s *ClassicPart) Name() string {
	return "classic"
}

// Origin returns the origin ("")
func (s *ClassicPart) Origin() string {
	return ""
}

// Version returns the version
func (s *ClassicPart) Version() string {
	return "fake-me"
}

// Description returns the description
func (s *ClassicPart) Description() string {
	return i18n.G("The ubuntu classic environment")
}

// Hash returns the hash
func (s *ClassicPart) Hash() string {
	return ""
}

// IsActive returns true if the snap is active
func (s *ClassicPart) IsActive() bool {
	return classic.Enabled()
}

// IsInstalled returns true if the snap is installed
func (s *ClassicPart) IsInstalled() bool {
	return classic.Enabled()
}

// InstalledSize returns the size of the installed snap
func (s *ClassicPart) InstalledSize() int64 {
	return -1
}

// DownloadSize returns the dowload size
func (s *ClassicPart) DownloadSize() int64 {
	return 0
}

// Date returns the last update date
func (s *ClassicPart) Date() time.Time {
	st, err := os.Stat(filepath.Join(dirs.ClassicDir, "/var/lib/dpkg/status"))
	if err != nil {
		// nothing meaningful
		return time.Now()
	}

	return st.ModTime()
}

// SetActive sets the snap active
func (s *ClassicPart) SetActive(active bool, pb progress.Meter) error {
	return nil
}

// Install installs the snap
func (s *ClassicPart) Install(pb progress.Meter, flags InstallFlags) (name string, err error) {
	// FIXME: pass in the pbar
	//return err = classic.Create(pbar)
	err = classic.Create()
	return "classic", err
}

// Uninstall can not be used for "core" snaps
func (s *ClassicPart) Uninstall(progress.Meter) error {
	return classic.Destroy()
}

// Config is used to to configure the snap
func (s *ClassicPart) Config(configuration []byte) (newConfig string, err error) {
	return "", nil
}

// NeedsReboot returns true if the snap becomes active on the next reboot
func (s *ClassicPart) NeedsReboot() bool {

	return false
}

// Channel returns the system-image-server channel used
func (s *ClassicPart) Channel() string {
	return "stable"
}

// Icon returns the icon path
func (s *ClassicPart) Icon() string {
	return ""
}

// Frameworks returns the list of frameworks needed by the snap
func (s *ClassicPart) Frameworks() ([]string, error) {
	// system image parts can't depend on frameworks.
	return nil, nil
}

// ClassicRepository is the type used for the system-image-server
type ClassicRepository struct {
}

// NewSystemImageRepository returns a new SystemImageRepository
func NewClassicRepository() *ClassicRepository {
	return &ClassicRepository{}
}

// Description describes the repository
func (s *ClassicRepository) Description() string {
	return "ClassicRepository"
}

// Search searches the ClassicRepository for the given terms
func (s *ClassicRepository) Search(terms string) (versions []Part, err error) {
	if strings.Contains(terms, "classic") {
		part := &ClassicPart{}
		versions = append(versions, part)
	}

	return versions, err
}

// Details returns details for the given snap
func (s *ClassicRepository) Details(name string, origin string) ([]Part, error) {
	if name == "classic" && origin == "" {
		return []Part{&ClassicPart{}}, nil
	}

	return nil, ErrPackageNotFound
}

// Updates returns the available updates
func (s *ClassicRepository) Updates() ([]Part, error) {
	return nil, nil
}

// Installed returns the installed snaps from this repository
func (s *ClassicRepository) Installed() (parts []Part, err error) {
	if !classic.Enabled() {
		return nil, nil
	}

	parts = append(parts, &ClassicPart{})
	return parts, nil
}

// All installed parts.
func (s *ClassicRepository) All() ([]Part, error) {
	return s.Installed()
}
