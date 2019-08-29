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

package patch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/snapcore/snapd/overlord/state"
)

type patch63auxStoreInfo struct {
	Media interface{} `json:"media,omitempty"`
}

func normChan(in string) string {
	return strings.Join(strings.FieldsFunc(in, func(r rune) bool { return r == '/' }), "/")
}

// patch6_3:
//  - ensure channel spec is valid
func patch6_3(st *state.State) error {
	var snaps map[string]*json.RawMessage
	if err := st.Get("snaps", &snaps); err != nil && err != state.ErrNoState {
		return fmt.Errorf("internal error: cannot get snaps: %s", err)
	}

	// Migrate snapstate
	dirty := false
	for name, raw := range snaps {
		var snapst map[string]interface{}
		if err := json.Unmarshal([]byte(*raw), &snapst); err != nil {
			return err
		}
		ch := snapst["channel"].(string)
		if ch != "" {
			normed := normChan(ch)
			if normed != ch {
				snapst["channel"] = normed
				data, err := json.Marshal(snapst)
				if err != nil {
					return err
				}
				newRaw := json.RawMessage(data)
				snaps[name] = &newRaw
				dirty = true
			}
		}
	}
	if dirty {
		st.Set("snaps", snaps)
	}

	// migrate tasks' snap setup
	for _, task := range st.Tasks() {
		chg := task.Change()
		if chg != nil && chg.Status().Ready() {
			continue
		}

		var snapsup map[string]interface{}
		err := task.Get("snap-setup", &snapsup)
		if err != nil && err != state.ErrNoState {
			return fmt.Errorf("internal error: cannot get snap-setup of task %s: %s", task.ID(), err)
		}
		if err == nil {
			ch := snapsup["channel"].(string)
			normed := normChan(ch)
			if normed != ch {
				snapsup["channel"] = normed
				task.Set("snap-setup", snapsup)
			}
			task.Get("old-channel", &ch)
			normed = normChan(ch)
			if normed != ch {
				task.Set("old-channel", normed)
			}
		}
	}

	return nil
}
