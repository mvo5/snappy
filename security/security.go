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

package security

import (
	"fmt"
	"strings"

	"github.com/ubuntu-core/snappy/snap"
)

// Profile returns the security profile string in the form of
// "snap_app-path_version" or an error
func Profile(m *snap.Info, appName, baseDir string) (string, error) {
	cleanedName := strings.Replace(appName, "/", "-", -1)
	if m.Type == snap.TypeFramework || m.Type == snap.TypeGadget {
		return fmt.Sprintf("%s_%s_%s", m.Name, cleanedName, m.Version), nil
	}

	origin, err := originFromBasedir(baseDir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s_%s_%s", m.Name, origin, cleanedName, m.Version), nil
}
