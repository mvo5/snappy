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

package errtracker

import (
	"time"
)

func MockCrashDbURL(url string) (restorer func()) {
	old := crashDbUrlBase
	crashDbUrlBase = url
	return func() {
		crashDbUrlBase = old
	}
}

func MockTimeNow(f func() time.Time) (restorer func()) {
	old := timeNow
	timeNow = f
	return func() {
		timeNow = old
	}
}
