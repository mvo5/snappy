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
	"github.com/snapcore/snapd/snap"
)

type patch63Flags struct {
	DevMode          bool `json:"devmode,omitempty"`
	JailMode         bool `json:"jailmode,omitempty"`
	Classic          bool `json:"classic,omitempty"`
	TryMode          bool `json:"trymode,omitempty"`
	Revert           bool `json:"revert,omitempty"`
	RemoveSnapPath   bool `json:"remove-snap-path,omitempty"`
	IgnoreValidation bool `json:"ignore-validation,omitempty"`
	Required         bool `json:"required,omitempty"`
	SkipConfigure    bool `json:"skip-configure,omitempty"`
	Unaliased        bool `json:"unaliased,omitempty"`
	Amend            bool `json:"amend,omitempty"`
	IsAutoRefresh    bool `json:"is-auto-refresh,omitempty"`
	NoReRefresh      bool `json:"no-rerefresh,omitempty"`
	RequireTypeBase  bool `json:"require-base-type,omitempty"`
}

type patch63SnapSetup struct {
	// FIXME: rename to RequestedChannel to convey the meaning better
	Channel string    `json:"channel,omitempty"`
	UserID  int       `json:"user-id,omitempty"`
	Base    string    `json:"base,omitempty"`
	Type    snap.Type `json:"type,omitempty"`
	// PlugsOnly indicates whether the relevant revisions for the
	// operation have only plugs (#plugs >= 0), and absolutely no
	// slots (#slots == 0).
	PlugsOnly bool `json:"plugs-only,omitempty"`

	CohortKey string `json:"cohort-key,omitempty"`

	// FIXME: implement rename of this as suggested in
	//  https://github.com/snapcore/snapd/pull/4103#discussion_r169569717
	//
	// Prereq is a list of snap-names that need to get installed
	// together with this snap. Typically used when installing
	// content-snaps with default-providers.
	Prereq []string `json:"prereq,omitempty"`

	patch63Flags

	SnapPath string `json:"snap-path,omitempty"`

	DownloadInfo *snap.DownloadInfo `json:"download-info,omitempty"`
	SideInfo     *snap.SideInfo     `json:"side-info,omitempty"`
	patch63auxStoreInfo

	// InstanceKey is set by the user during installation and differs for
	// each instance of given snap
	InstanceKey string `json:"instance-key,omitempty"`
}

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

		var snapsup patch63SnapSetup
		err := task.Get("snap-setup", &snapsup)
		if err != nil && err != state.ErrNoState {
			return fmt.Errorf("internal error: cannot get snap-setup of task %s: %s", task.ID(), err)
		}
		if err == nil {
			ch := snapsup.Channel
			normed := normChan(ch)
			if normed != ch {
				snapsup.Channel = normed
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
