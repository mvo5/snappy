// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2018 Canonical Ltd
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
	"sort"

	"golang.org/x/net/context"

	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/store"
	"github.com/snapcore/snapd/strutil"
)

type updateInfoOpts struct {
	channel          string
	ignoreValidation bool
	amend            bool
}

func idForUser(user *auth.UserState) int {
	if user == nil {
		return 0
	}
	return user.ID
}

func userIDForSnap(st *state.State, snapst *SnapState, fallbackUserID int) (int, error) {
	userID := snapst.UserID
	_, err := auth.User(st, userID)
	if err == nil {
		return userID, nil
	}
	if err != auth.ErrInvalidUser {
		return 0, err
	}
	return fallbackUserID, nil
}

// userFromUserID returns the first valid user from a series of userIDs
// used as successive fallbacks.
func userFromUserID(st *state.State, userIDs ...int) (*auth.UserState, error) {
	var user *auth.UserState
	var err error
	for _, userID := range userIDs {
		if userID == 0 {
			err = nil
			continue
		}
		user, err = auth.User(st, userID)
		if err != auth.ErrInvalidUser {
			break
		}
	}
	return user, err
}

// userFromUserIDOrFallback returns the user corresponding to userID
// if valid or otherwise the fallbackUser.
func userFromUserIDOrFallback(st *state.State, userID int, fallbackUser *auth.UserState) (*auth.UserState, error) {
	if userID != 0 {
		u, err := auth.User(st, userID)
		if err != nil && err != auth.ErrInvalidUser {
			return nil, err
		}
		if err == nil {
			return u, nil
		}
	}
	return fallbackUser, nil
}

func installInfo(st *state.State, name, channel string, revision snap.Revision, userID int) (*snap.Info, error) {
	// TODO: support ignore-validation?

	curSnaps, err := currentSnaps(st)
	if err != nil {
		return nil, err
	}

	user, err := userFromUserID(st, userID)
	if err != nil {
		return nil, err
	}

	// cannot specify both with the API
	if !revision.Unset() {
		channel = ""
	}

	action := &store.SnapAction{
		Action: "install",
		Name:   name,
		// the desired channel
		Channel: channel,
		// the desired revision
		Revision: revision,
	}

	theStore := Store(st)
	st.Unlock() // calls to the store should be done without holding the state lock
	res, err := theStore.SnapAction(context.TODO(), curSnaps, []*store.SnapAction{action}, user, nil)
	st.Lock()

	return singleActionResult(name, action.Action, res, err)
}

func updateInfo(st *state.State, snapst *SnapState, opts *updateInfoOpts, userID int) (*snap.Info, error) {
	if opts == nil {
		opts = &updateInfoOpts{}
	}

	curSnaps, err := currentSnaps(st)
	if err != nil {
		return nil, err
	}

	curInfo, user, err := preUpdateInfo(st, snapst, opts.amend, userID)
	if err != nil {
		return nil, err
	}

	var flags store.SnapActionFlags
	if opts.ignoreValidation {
		flags = store.SnapActionIgnoreValidation
	} else {
		flags = store.SnapActionEnforceValidation
	}

	action := &store.SnapAction{
		Action: "refresh",
		SnapID: curInfo.SnapID,
		// the desired channel
		Channel: opts.channel,
		Flags:   flags,
	}

	if curInfo.SnapID == "" { // amend
		action.Action = "install"
		action.Name = curInfo.Name()
	}

	theStore := Store(st)
	st.Unlock() // calls to the store should be done without holding the state lock
	res, err := theStore.SnapAction(context.TODO(), curSnaps, []*store.SnapAction{action}, user, nil)
	st.Lock()

	return singleActionResult(curInfo.Name(), action.Action, res, err)
}

func preUpdateInfo(st *state.State, snapst *SnapState, amend bool, userID int) (*snap.Info, *auth.UserState, error) {
	user, err := userFromUserID(st, snapst.UserID, userID)
	if err != nil {
		return nil, nil, err
	}

	curInfo, err := snapst.CurrentInfo()
	if err != nil {
		return nil, nil, err
	}

	if curInfo.SnapID == "" { // covers also trymode
		if !amend {
			return nil, nil, store.ErrLocalSnap
		}
	}

	return curInfo, user, nil
}

func singleActionResult(name, action string, results []*snap.Info, e error) (info *snap.Info, err error) {
	if len(results) > 1 {
		return nil, fmt.Errorf("internal error: multiple store results for a single snap op")
	}
	if len(results) > 0 {
		// TODO: if we also have an error log/warn about it
		return results[0], nil
	}

	if saErr, ok := e.(*store.SnapActionError); ok {
		if len(saErr.Other) != 0 {
			return nil, saErr
		}

		var snapErr error
		switch action {
		case "refresh":
			snapErr = saErr.Refresh[name]
		case "install":
			snapErr = saErr.Install[name]
		}
		if snapErr != nil {
			return nil, snapErr
		}

		// no result, atypical case
		if saErr.NoResults {
			switch action {
			case "refresh":
				return nil, store.ErrNoUpdateAvailable
			case "install":
				return nil, store.ErrSnapNotFound
			}
		}
	}

	return nil, e
}

func updateToRevisionInfo(st *state.State, snapst *SnapState, revision snap.Revision, userID int) (*snap.Info, error) {
	// TODO: support ignore-validation?

	curSnaps, err := currentSnaps(st)
	if err != nil {
		return nil, err
	}

	curInfo, user, err := preUpdateInfo(st, snapst, false, userID)
	if err != nil {
		return nil, err
	}

	action := &store.SnapAction{
		Action: "refresh",
		SnapID: curInfo.SnapID,
		// the desired revision
		Revision: revision,
	}

	theStore := Store(st)
	st.Unlock() // calls to the store should be done without holding the state lock
	res, err := theStore.SnapAction(context.TODO(), curSnaps, []*store.SnapAction{action}, user, nil)
	st.Lock()

	return singleActionResult(curInfo.Name(), action.Action, res, err)
}

func currentSnaps(st *state.State) ([]*store.CurrentSnap, error) {
	snapStates, err := All(st)
	if err != nil {
		return nil, err
	}

	curSnaps := collectCurrentSnaps(snapStates, nil)
	return curSnaps, nil
}

func collectCurrentSnaps(snapStates map[string]*SnapState, consider func(*store.CurrentSnap, *SnapState)) (curSnaps []*store.CurrentSnap) {
	curSnaps = make([]*store.CurrentSnap, 0, len(snapStates))

	for snapName, snapst := range snapStates {
		if snapst.TryMode {
			// try mode snaps are completely local and
			// irrelevant for the operation
			continue
		}

		snapInfo, err := snapst.CurrentInfo()
		if err != nil {
			continue
		}

		if snapInfo.SnapID == "" {
			// the store won't be able to tell what this
			// is and so cannot include it in the
			// operation
			continue
		}

		installed := &store.CurrentSnap{
			Name:   snapName,
			SnapID: snapInfo.SnapID,
			// the desired channel (not snapInfo.Channel!)
			TrackingChannel:  snapst.Channel,
			Revision:         snapInfo.Revision,
			RefreshedDate:    revisionDate(snapInfo),
			IgnoreValidation: snapst.IgnoreValidation,
		}
		curSnaps = append(curSnaps, installed)

		if consider != nil {
			consider(installed, snapst)
		}
	}

	return curSnaps
}

func refreshCandidates(ctx context.Context, st *state.State, names []string, user *auth.UserState, opts *store.RefreshOptions) ([]*snap.Info, map[string]*SnapState, map[string]bool, error) {
	snapStates, err := All(st)
	if err != nil {
		return nil, nil, nil, err
	}

	// check if we have this name at all
	for _, name := range names {
		if _, ok := snapStates[name]; !ok {
			return nil, nil, nil, snap.NotInstalledError{Snap: name}
		}
	}

	sort.Strings(names)

	actionsByUserID := make(map[int][]*store.SnapAction)
	stateByID := make(map[string]*SnapState, len(snapStates))
	ignoreValidation := make(map[string]bool)
	fallbackID := idForUser(user)
	nCands := 0

	addCand := func(installed *store.CurrentSnap, snapst *SnapState) {
		// FIXME: snaps that are not active are skipped for now
		//        until we know what we want to do
		if !snapst.Active {
			return
		}

		if len(names) == 0 && snapst.DevMode {
			// no auto-refresh for devmode
			return
		}

		if len(names) > 0 && !strutil.SortedListContains(names, installed.Name) {
			return
		}

		stateByID[installed.SnapID] = snapst

		if len(names) == 0 {
			installed.Block = snapst.Block()
		}

		userID := snapst.UserID
		if userID == 0 {
			userID = fallbackID
		}
		actionsByUserID[userID] = append(actionsByUserID[userID], &store.SnapAction{
			Action: "refresh",
			SnapID: installed.SnapID,
		})
		if snapst.IgnoreValidation {
			ignoreValidation[installed.SnapID] = true
		}
		nCands++
	}
	// determine current snaps and collect candidates for refresh
	curSnaps := collectCurrentSnaps(snapStates, addCand)

	theStore := Store(st)

	updatesInfo := make(map[string]*snap.Info, nCands)
	for userID, actions := range actionsByUserID {
		u, err := userFromUserIDOrFallback(st, userID, user)
		if err != nil {
			return nil, nil, nil, err
		}

		st.Unlock()
		updatesForUser, err := theStore.SnapAction(ctx, curSnaps, actions, u, opts)
		st.Lock()
		if err != nil {
			saErr, ok := err.(*store.SnapActionError)
			if !ok {
				return nil, nil, nil, err
			}
			// TODO: use the warning infra here when we have it
			logger.Noticef("%v", saErr)
		}

		for _, snapInfo := range updatesForUser {
			updatesInfo[snapInfo.SnapID] = snapInfo
		}
	}

	updates := make([]*snap.Info, 0, len(updatesInfo))
	for _, snapInfo := range updatesInfo {
		updates = append(updates, snapInfo)
	}

	return updates, stateByID, ignoreValidation, nil
}
