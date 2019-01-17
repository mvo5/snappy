// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
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

package timeutil

import (
	"fmt"
	"time"

	"github.com/snapcore/snapd/logger"
)

type Measure struct {
	action     string
	start, end time.Time
}

func NewMeasure(action string) *Measure {
	return &Measure{action: action, start: time.Now()}
}

func (m *Measure) LogDone() {
	if m.end.IsZero() {
		m.end = time.Now()
	}
	msg := fmt.Sprintf("%s took %v", m.action, m.end.Sub(m.start))
	logger.Noticeff("perf: %s", msg)
}
