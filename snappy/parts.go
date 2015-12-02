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
	"fmt"
	"net"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/parts/part"
	"github.com/ubuntu-core/snappy/pkg"
	"github.com/ubuntu-core/snappy/progress"
)

// SystemConfig is a config map holding configs for multiple packages
type SystemConfig map[string]interface{}

// ServiceYamler implements snappy packages that offer services
type ServiceYamler interface {
	ServiceYamls() []ServiceYaml
}

// Configuration allows requesting an oem snappy package type's config
type Configuration interface {
	OemConfig() SystemConfig
}

// QualifiedName of a part.IF is the Name, in most cases qualified with the
// Origin
func QualifiedName(p part.IF) string {
	if t := p.Type(); t == pkg.TypeFramework || t == pkg.TypeOem {
		return p.Name()
	}
	return p.Name() + "." + p.Origin()
}

// BareName of a part.IF is just its Name
func BareName(p part.IF) string {
	return p.Name()
}

// FullName of a part.IF is Name.Origin
func FullName(p part.IF) string {
	return p.Name() + "." + p.Origin()
}

// FullNameWithChannel returns the FullName, with the channel appended
// if it has one.
func fullNameWithChannel(p part.IF) string {
	name := FullName(p)
	ch := p.Channel()
	if ch == "" {
		return name
	}

	return fmt.Sprintf("%s/%s", name, ch)
}

// Repository is the interface for a collection of snaps
type Repository interface {

	// query
	Description() string

	// action
	Details(name string, origin string) ([]part.IF, error)

	Updates() ([]part.IF, error)
	Installed() ([]part.IF, error)

	All() ([]part.IF, error)
}

// MetaRepository contains all available single repositories can can be used
// to query in a single place
type MetaRepository struct {
	all []Repository
}

// NewMetaStoreRepository returns a MetaRepository of stores
func NewMetaStoreRepository() *MetaRepository {
	m := new(MetaRepository)
	m.all = []Repository{}

	if repo := NewUbuntuStoreSnapRepository(); repo != nil {
		m.all = append(m.all, repo)
	}

	return m
}

// NewMetaLocalRepository returns a MetaRepository of stores
func NewMetaLocalRepository() *MetaRepository {
	m := new(MetaRepository)
	m.all = []Repository{}

	if repo := NewSystemImageRepository(); repo != nil {
		m.all = append(m.all, repo)
	}
	if repo := NewLocalSnapRepository(dirs.SnapAppsDir); repo != nil {
		m.all = append(m.all, repo)
	}
	if repo := NewLocalSnapRepository(dirs.SnapOemDir); repo != nil {
		m.all = append(m.all, repo)
	}

	return m
}

// NewMetaRepository returns a new MetaRepository
func NewMetaRepository() *MetaRepository {
	// FIXME: make this a configuration file

	m := NewMetaLocalRepository()
	if repo := NewUbuntuStoreSnapRepository(); repo != nil {
		m.all = append(m.all, repo)
	}

	return m
}

// Installed returns all installed parts
func (m *MetaRepository) Installed() (parts []part.IF, err error) {
	for _, r := range m.all {
		installed, err := r.Installed()
		if err != nil {
			return parts, err
		}
		parts = append(parts, installed...)
	}

	return parts, err
}

// All the parts
func (m *MetaRepository) All() ([]part.IF, error) {
	var parts []part.IF

	for _, r := range m.all {
		all, err := r.All()
		if err != nil {
			return nil, err
		}
		parts = append(parts, all...)
	}

	return parts, nil
}

// Updates returns all updatable parts
func (m *MetaRepository) Updates() (parts []part.IF, err error) {
	for _, r := range m.all {
		updates, err := r.Updates()
		if err != nil {
			return parts, err
		}
		parts = append(parts, updates...)
	}

	return parts, err
}

// Details returns details for the given snap name
func (m *MetaRepository) Details(name string, origin string) ([]part.IF, error) {
	var parts []part.IF

	for _, r := range m.all {
		results, err := r.Details(name, origin)
		// ignore network errors here, we will also collect
		// local results
		_, netError := err.(net.Error)
		_, urlError := err.(*url.Error)
		switch {
		case err == ErrPackageNotFound || netError || urlError:
			continue
		case err != nil:
			return nil, err
		}
		parts = append(parts, results...)
	}

	return parts, nil
}

// ActiveSnapsByType returns all installed snaps with the given type
func ActiveSnapsByType(snapTs ...pkg.Type) (res []part.IF, err error) {
	m := NewMetaRepository()
	installed, err := m.Installed()
	if err != nil {
		return nil, err
	}

	for _, part := range installed {
		if !part.IsActive() {
			continue
		}
		for i := range snapTs {
			if part.Type() == snapTs[i] {
				res = append(res, part)
			}
		}
	}

	return res, nil
}

// ActiveSnapIterByType returns the result of applying the given
// function to all active snaps with the given type.
var ActiveSnapIterByType = activeSnapIterByTypeImpl

func activeSnapIterByTypeImpl(f func(part.IF) string, snapTs ...pkg.Type) ([]string, error) {
	installed, err := ActiveSnapsByType(snapTs...)
	res := make([]string, len(installed))

	for i, part := range installed {
		res[i] = f(part)
	}

	return res, err
}

// ActiveSnapByName returns all active snaps with the given name
func ActiveSnapByName(needle string) part.IF {
	m := NewMetaRepository()
	installed, err := m.Installed()
	if err != nil {
		return nil
	}
	for _, part := range installed {
		if !part.IsActive() {
			continue
		}
		if part.Name() == needle {
			return part
		}
	}

	return nil
}

// FindSnapsByName returns all snaps with the given name in the "haystack"
// slice of parts (useful for filtering)
func FindSnapsByName(needle string, haystack []part.IF) (res []part.IF) {
	name, origin := SplitOrigin(needle)
	ignorens := origin == ""

	for _, part := range haystack {
		if part.Name() == name && (ignorens || part.Origin() == origin) {
			res = append(res, part)
		}
	}

	return res
}

// SplitOrigin splits a snappy name name into a (name, origin) pair
func SplitOrigin(name string) (string, string) {
	idx := strings.LastIndexAny(name, ".")
	if idx > -1 {
		return name[:idx], name[idx+1:]
	}

	return name, ""
}

// FindSnapsByNameAndVersion returns the parts with the name/version in the
// given slice of parts
func FindSnapsByNameAndVersion(needle, version string, haystack []part.IF) []part.IF {
	name, origin := SplitOrigin(needle)
	ignorens := origin == ""
	var found []part.IF

	for _, part := range haystack {
		if part.Name() == name && part.Version() == version && (ignorens || part.Origin() == origin) {
			found = append(found, part)
		}
	}

	return found
}

// MakeSnapActiveByNameAndVersion makes the given snap version the active
// version
func makeSnapActiveByNameAndVersion(pkg, ver string, inter progress.Meter) error {
	m := NewMetaRepository()
	installed, err := m.Installed()
	if err != nil {
		return err
	}

	parts := FindSnapsByNameAndVersion(pkg, ver, installed)
	switch len(parts) {
	case 0:
		return fmt.Errorf("Can not find %s with version %s", pkg, ver)
	case 1:
		return parts[0].SetActive(true, inter)
	default:
		return fmt.Errorf("More than one %s with version %s", pkg, ver)
	}
}

// PackageNameActive checks whether a fork of the given name is active in the system
func PackageNameActive(name string) bool {
	return ActiveSnapByName(name) != nil
}

// iconPath returns the would be path for the local icon
func iconPath(s part.IF) string {
	// TODO: care about extension ever being different than png
	return filepath.Join(dirs.SnapIconsDir, fmt.Sprintf("%s_%s.png", QualifiedName(s), s.Version()))
}

// RemoteManifestPath returns the would be path for the store manifest meta data
func RemoteManifestPath(s part.IF) string {
	return filepath.Join(dirs.SnapMetaDir, fmt.Sprintf("%s_%s.manifest", QualifiedName(s), s.Version()))
}
