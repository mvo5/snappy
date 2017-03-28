// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package osutil_test

import (
	"github.com/snapcore/snapd/osutil"
)

var truePath string
var falsePath string
var gccPath string

func init() {
	truePath = osutil.FindInPathOrDefault("true", "/bin/true")
	falsePath = osutil.FindInPathOrDefault("false", "/bin/false")
	gccPath = osutil.FindInPathOrDefault("gcc", "/usr/bin/gcc")
}
