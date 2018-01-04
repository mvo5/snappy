// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
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

package snapstate

import (
	"fmt"

	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/store"
)

func userFromUserID(st *state.State, userID int) (*auth.UserState, error) {
	if userID == 0 {
		return nil, nil
	}
	return auth.User(st, userID)
}

func snapNameToID(st *state.State, name string, user *auth.UserState) (string, error) {
	theStore := Store(st)
	st.Unlock()
	info, err := theStore.SnapInfo(store.SnapSpec{Name: name}, user)
	st.Lock()
	return info.SnapID, err
}

func updateInfo(st *state.State, snapst *SnapState, channel string, ignoreValidation, amend bool, userID int) (*snap.Info, error) {
	user, err := userFromUserID(st, userID)
	if err != nil {
		return nil, err
	}
	curInfo, err := snapst.CurrentInfo()
	if err != nil {
		return nil, err
	}

	if curInfo.SnapID == "" { // covers also trymode
		if !amend {
			return nil, store.ErrLocalSnap
		}

		// in amend mode we need to move to the store rev
		id, err := snapNameToID(st, curInfo.Name(), user)
		if err != nil {
			return nil, fmt.Errorf("cannot get snap ID for %q: %v", curInfo.Name(), err)
		}
		curInfo.SnapID = id
		// set revision to "unknown"
		curInfo.Revision = snap.R(0)
	}

	refreshCand := &store.RefreshCandidate{
		// the desired channel
		Channel:          channel,
		SnapID:           curInfo.SnapID,
		Revision:         curInfo.Revision,
		Epoch:            curInfo.Epoch,
		IgnoreValidation: ignoreValidation,
		Amend:            amend,
	}

	theStore := Store(st)
	st.Unlock() // calls to the store should be done without holding the state lock
	res, err := theStore.LookupRefresh(refreshCand, user)
	st.Lock()
	return res, err
}

func snapInfo(st *state.State, name, channel string, revision snap.Revision, userID int) (*snap.Info, error) {
	user, err := userFromUserID(st, userID)
	if err != nil {
		return nil, err
	}
	theStore := Store(st)
	st.Unlock() // calls to the store should be done without holding the state lock
	spec := store.SnapSpec{
		Name:     name,
		Channel:  channel,
		Revision: revision,
	}
	snap, err := theStore.SnapInfo(spec, user)
	st.Lock()
	return snap, err
}
