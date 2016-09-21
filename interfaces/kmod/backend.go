// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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

package kmod

import (
	"bytes"
	"fmt"

	"github.com/snapcore/snapd/interfaces"
	"github.com/snapcore/snapd/snap"
)

// Backend is responsible for maintaining kernel modules
type Backend struct{}

// Name returns the name of the backend.
func (b *Backend) Name() string {
	return "kmod"
}

// Setup creates a conf file with list of kernel modules required by given snap,
// writes it in /etc/modules-load.d/ directory and immediately loads the modules
// using /sbin/modprobe. The devMode is ignored.
//
// If the method fails it should be re-tried (with a sensible strategy) by the caller.
func (b *Backend) Setup(snapInfo *snap.Info, devMode bool, repo *interfaces.Repository) error {
	snapName := snapInfo.Name()
	// Get the snippets that apply to this snap
	snippets, err := repo.SecuritySnippetsForSnap(snapInfo.Name(), interfaces.SecurityKMod)
	if err != nil {
		return fmt.Errorf("cannot obtain kmod security snippets for snap %q: %s", snapName, err)
	}
	if len(snippets) == 0 {
		// Make sure that the modules conf file gets removed when we don't have any content
		return removeModulesFile(snapName)
	}

	modules := b.processSnipets(snapInfo, snippets)
	err = loadModules(modules)
	if err == nil {
		err = writeModulesFile(modules, snapName)
	}
	if err != nil {
		return err
	}
	return nil
}

// Remove removes modules config file specific to a given snap.
//
// This method should be called after removing a snap.
//
// If the method fails it should be re-tried (with a sensible strategy) by the caller.
func (b *Backend) Remove(snapName string) error {
	removeModulesFile(snapName)
	return nil
}

// processSnipets combines security snippets collected from all the interfaces
// affecting a given snap into a de-duplicated list of kernel modules.
func (b *Backend) processSnipets(snapInfo *snap.Info, snippets map[string][][]byte) (modules [][]byte) {
	// we need to de-duplicate the modules, as some interfaces may contain overlapping modules.
	modulesDedup := make(map[string]struct{})
	for _, appInfo := range snapInfo.Apps {
		for _, snippet := range snippets[appInfo.SecurityTag()] {
			// split snippet by newline to get the list of modules
			individualLines := bytes.Split(snippet, []byte{'\n'})
			for _, line := range individualLines {
				l := bytes.Trim(line, " \r")
				// ignore empty lines and comments
				if len(l) > 0 && l[0] != '#' {
					modulesDedup[string(l)] = struct{}{}
				}
			}
		}
	}
	for mod, _ := range modulesDedup {
		modules = append(modules, []byte(mod))
	}
	return modules
}
