// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package daemon

import (
	"encoding/json"
	"net/http"

	"golang.org/x/net/context"
	"gopkg.in/check.v1"

	"github.com/snapcore/snapd/overlord"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/overlord/snapstate"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/snap"
)

var (
	SnapCmd    = snapCmd
	FindCmd    = findCmd
	RootCmd    = rootCmd
	SysInfoCmd = sysInfoCmd
	LoginCmd   = loginCmd

	LoginUser   = loginUser
	SearchStore = searchStore

	MapLocal = mapLocal
	Api      = api

	GetSnapInfo  = getSnapInfo
	GetSnapsInfo = getSnapsInfo
)

func (d *Daemon) Overlord() *overlord.Overlord {
	return d.overlord
}

func (d *Daemon) SetOverlord(o *overlord.Overlord) {
	d.overlord = o
}

func (d *Daemon) AddRoutes() {
	d.addRoutes()
}

func GetResp(c *check.C, respo Response) *resp {
	rsp, ok := respo.(*resp)
	c.Assert(ok, check.Equals, true)
	return rsp
}

func NewResp() *resp {
	return &resp{}
}

func UnmarshalResp(b []byte) (*resp, error) {
	var rsp resp
	err := json.Unmarshal(b, &rsp)
	return &rsp, err
}

func NewAboutSnap(info *snap.Info, snapst *snapstate.SnapState) aboutSnap {
	return aboutSnap{info: info, snapst: snapst}
}

func MockMuxVars(f func(*http.Request) map[string]string) (restore func()) {
	oldMuxVars := muxVars
	muxVars = f
	return func() {
		muxVars = oldMuxVars
	}
}

func MockAssertstateRefreshSnapDeclarations(f func(s *state.State, userID int) error) (restore func()) {
	oldAssertstateRefreshSnapDeclarations := assertstateRefreshSnapDeclarations
	assertstateRefreshSnapDeclarations = f
	return func() {
		assertstateRefreshSnapDeclarations = oldAssertstateRefreshSnapDeclarations
	}
}

func MockSnapstateInstall(f func(st *state.State, name, channel string, revision snap.Revision, userID int, flags snapstate.Flags) (*state.TaskSet, error)) (restore func()) {
	oldSnapstateInstall := snapstateInstall
	snapstateInstall = f
	return func() {
		snapstateInstall = oldSnapstateInstall
	}
}

func MockSnapstateInstallPath(f func(st *state.State, si *snap.SideInfo, path, instanceName, channel string, flags snapstate.Flags) (*state.TaskSet, *snap.Info, error)) (restore func()) {
	oldSnapstateInstallPath := snapstateInstallPath
	snapstateInstallPath = f
	return func() {
		snapstateInstallPath = oldSnapstateInstallPath
	}
}

func MockSnapstateInstallMany(f func(st *state.State, names []string, userID int) ([]string, []*state.TaskSet, error)) (restore func()) {
	oldSnapstateInstallMany := snapstateInstallMany
	snapstateInstallMany = f
	return func() {
		snapstateInstallMany = oldSnapstateInstallMany
	}
}

func MockSnapstateRefreshCandidates(f func(st *state.State, user *auth.UserState) ([]*snap.Info, error)) (restore func()) {
	oldSnapstateRefreshCandidates := snapstateRefreshCandidates
	snapstateRefreshCandidates = f
	return func() {
		snapstateRefreshCandidates = oldSnapstateRefreshCandidates
	}
}

func MockSnapstateRemoveMany(f func(st *state.State, names []string) ([]string, []*state.TaskSet, error)) (restore func()) {
	oldSnapstateRemoveMany := snapstateRemoveMany
	snapstateRemoveMany = f
	return func() {
		snapstateRemoveMany = oldSnapstateRemoveMany
	}
}

func MockSnapstateRevert(f func(st *state.State, name string, flags snapstate.Flags) (*state.TaskSet, error)) (restore func()) {
	oldSnapstateRevert := snapstateRevert
	snapstateRevert = f
	return func() {
		snapstateRevert = oldSnapstateRevert
	}
}

func MockSnapstateRevertToRevision(f func(st *state.State, name string, rev snap.Revision, flags snapstate.Flags) (*state.TaskSet, error)) (restore func()) {
	oldSnapstateRevertToRevision := snapstateRevertToRevision
	snapstateRevertToRevision = f
	return func() {
		snapstateRevertToRevision = oldSnapstateRevertToRevision
	}
}

func MockTryPath(f func(st *state.State, name, path string, flags snapstate.Flags) (*state.TaskSet, error)) (restore func()) {
	oldSnapstateTryPath := snapstateTryPath
	snapstateTryPath = snapstateTryPath
	return func() {
		snapstateTryPath = oldSnapstateTryPath
	}
}

func MockUpdate(f func(st *state.State, name, channel string, revision snap.Revision, userID int, flags snapstate.Flags) (*state.TaskSet, error)) (restore func()) {
	oldSnapstateUpdate := snapstateUpdate
	snapstateUpdate = snapstateUpdate
	return func() {
		snapstateUpdate = oldSnapstateUpdate
	}
}

func MockUpdateMany(f func(ctx context.Context, st *state.State, names []string, userID int) ([]string, []*state.TaskSet, error)) (restore func()) {
	oldSnapstateUpdateMany := snapstateUpdateMany
	snapstateUpdateMany = snapstateUpdateMany
	return func() {
		snapstateUpdateMany = oldSnapstateUpdateMany
	}
}

func MockUnsafeReadSnapInfo(f func(snapPath string) (*snap.Info, error)) (restore func()) {
	oldUnsafeReadSnapInfo := unsafeReadSnapInfo
	unsafeReadSnapInfo = f
	return func() {
		unsafeReadSnapInfo = oldUnsafeReadSnapInfo
	}
}

func MockEnsureStateSoon(f func(st *state.State)) (restore func()) {
	oldEnsureStateSoon := ensureStateSoon
	ensureStateSoon = f
	return func() {
		ensureStateSoon = oldEnsureStateSoon
	}
}
