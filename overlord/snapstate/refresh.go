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

package snapstate

import (
	"fmt"
	"sort"
	"strings"

	"github.com/snapcore/snapd/cmd/snaplock"
	"github.com/snapcore/snapd/cmd/snaplock/runinhibit"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/sandbox/cgroup"
	"github.com/snapcore/snapd/snap"
)

// pidsOfSnap is a mockable version of PidsOfSnap
var pidsOfSnap = cgroup.PidsOfSnap

var genericRefreshCheck = func(info *snap.Info, canAppRunDuringRefresh func(app *snap.AppInfo) bool) error {
	knownPids, err := pidsOfSnap(info.InstanceName())
	if err != nil {
		return err
	}

	// Due to specific of the interaction with locking, all locking is performed by the caller.
	var busyAppNames []string
	var busyHookNames []string
	var busyPIDs []int

	// Currently there are no situations when hooks might be allowed to run
	// during the refresh process. The function exists to make the next two
	// chunks of code symmetric.
	canHookRunDuringRefresh := func(hook *snap.HookInfo) bool {
		return false
	}

	for name, app := range info.Apps {
		if canAppRunDuringRefresh(app) {
			continue
		}
		if PIDs := knownPids[app.SecurityTag()]; len(PIDs) > 0 {
			busyAppNames = append(busyAppNames, name)
			busyPIDs = append(busyPIDs, PIDs...)
		}
	}

	for name, hook := range info.Hooks {
		if canHookRunDuringRefresh(hook) {
			continue
		}
		if PIDs := knownPids[hook.SecurityTag()]; len(PIDs) > 0 {
			busyHookNames = append(busyHookNames, name)
			busyPIDs = append(busyPIDs, PIDs...)
		}
	}
	if len(busyAppNames) == 0 && len(busyHookNames) == 0 {
		return nil
	}
	sort.Strings(busyAppNames)
	sort.Strings(busyHookNames)
	sort.Ints(busyPIDs)
	return &BusySnapError{
		SnapName:      info.SnapName(),
		busyAppNames:  busyAppNames,
		busyHookNames: busyHookNames,
		pids:          busyPIDs,
	}
}

// SoftNothingRunningRefreshCheck looks if there are at most only service processes alive.
//
// The check is designed to run early in the refresh pipeline. Before
// downloading or stopping services for the update, we can check that only
// services are running, that is, that no non-service apps or hooks are
// currently running.
//
// Since services are stopped during the update this provides a good early
// precondition check. The check is also deliberately racy as existing snap
// commands can fork new processes or existing processes can die. After the
// soft check passes the user is free to start snap applications and block the
// hard check.
func SoftNothingRunningRefreshCheck(info *snap.Info) error {
	return genericRefreshCheck(info, func(app *snap.AppInfo) bool {
		return app.IsService()
	})
}

// HardNothingRunningRefreshCheck looks if there are any undesired processes alive.
//
// The check is designed to run late in the refresh pipeline, after stopping
// snap services. At this point non-enduring services should be stopped, hooks
// should no longer run, and applications should be barred from running
// externally (e.g. by using a new inhibition mechanism for snap run).
//
// The check fails if any process belonging to the snap, apart from services
// that are enduring refresh, is still alive. If a snap is busy it cannot be
// refreshed and the refresh process is aborted.
func HardNothingRunningRefreshCheck(info *snap.Info) error {
	return genericRefreshCheck(info, func(app *snap.AppInfo) bool {
		// TODO: use a constant instead of "endure"
		return app.IsService() && app.RefreshMode == "endure"
	})
}

// BusySnapError indicates that snap has apps or hooks running and cannot refresh.
type BusySnapError struct {
	SnapName      string
	pids          []int
	busyAppNames  []string
	busyHookNames []string
}

// Error formats an error string describing what is running.
func (err *BusySnapError) Error() string {
	switch {
	case len(err.busyAppNames) > 0 && len(err.busyHookNames) > 0:
		return fmt.Sprintf("snap %q has running apps (%s) and hooks (%s)",
			err.SnapName, strings.Join(err.busyAppNames, ", "), strings.Join(err.busyHookNames, ", "))
	case len(err.busyAppNames) > 0:
		return fmt.Sprintf("snap %q has running apps (%s)",
			err.SnapName, strings.Join(err.busyAppNames, ", "))
	case len(err.busyHookNames) > 0:
		return fmt.Sprintf("snap %q has running hooks (%s)",
			err.SnapName, strings.Join(err.busyHookNames, ", "))
	default:
		return fmt.Sprintf("snap %q has running apps or hooks", err.SnapName)
	}
}

// Pids returns the set of process identifiers that are running.
//
// Since this list is a snapshot it should be only acted upon if there is an
// external synchronization system applied (e.g. all processes are frozen) at
// the time the snapshot was taken.
//
// The list is intended for snapd to forcefully kill all processes for a forced
// refresh scenario.
func (err BusySnapError) Pids() []int {
	return err.pids
}

// doHardRefreshCheck performs the complete hard refresh check interaction.
//
// This check embeds HardNothingRunningRefreshCheck along with interaction
// with two locks - the snap lock, shared by snap-confine and snapd and the
// snap run inhibition lock, shared by snapd and snap run.
//
// On success function returns a locked lock, allowing the caller to
// atomically, with regards to snap-confine, finish any action that required
// the check.
//
// In practice this is used to unlink the snap from disk. Once snap-confine
// resumes, the snap is no longer linked and normal startup is inhibited. In
// consequence "snap run foo" will block and snap-confine will fail with
// unlinked snap error.
func doHardRefreshCheck(st *state.State, snapst *SnapState, info *snap.Info) (lock *osutil.FileLock, err error) {
	// A process may be created after the soft refresh done upon
	// the request to refresh a snap. If such process is alive by
	// the time this code is reached the refresh process is stopped.

	// Grab per-snap lock to prevent new processes from starting. This is
	// sufficient to perform the check, even though individual processes
	// may fork or exit, we will have per-security-tag information about
	// what is running.
	lock, err = snaplock.OpenLock(info.InstanceName())
	if err != nil {
		return nil, err
	}
	defer func() {
		// If we have a lock but we are returning an error then unlock the lock
		// by closing it.
		if lock != nil && err != nil {
			lock.Close()
		}
	}()
	if err := lock.Lock(); err != nil {
		return nil, err
	}

	if err := inhibitRefresh(st, snapst, info, HardNothingRunningRefreshCheck); err != nil {
		// In case of successful inhibition the snap state is modified to
		// indicate when the refresh was first inhibited. If the first
		// inhibition is outside of a grace period then refresh proceeds
		// regardless of the existing processes.
		return nil, err
	}

	// Snap was not busy so we can refresh now. While we are still holding
	// the snap lock, obtain the run inhibition lock with a hint indicating
	// that refresh is in progress.

	// XXX: should we move this logic to the place that calls the "soft"
	// check instead? Doing so would somewhat change the semantic of soft
	// and hard checks, as it would effectively make hard check a no-op,
	// but it might provide a nicer user experience.
	if err := runinhibit.LockWithHint(info.InstanceName(), runinhibit.HintInhibitedForRefresh); err != nil {
		return nil, err
	}

	return lock, nil
}

// doSoftRefreshCheck performs the complete soft refresh check interaction.
//
// This check embeds SoftNothingRunningRefreshCheck along with interaction with
// the snap lock, shared by snap-confine and snapd.
func doSoftRefreshCheck(st *state.State, snapst *SnapState, info *snap.Info) error {
	// Grab per-snap lock to prevent new processes from starting. This is
	// sufficient to perform the check, even though individual processes may
	// fork or exit, we will have per-security-tag information about what is
	// running.
	lock, err := snaplock.OpenLock(info.InstanceName())
	if err != nil {
		return err
	}
	// Closing the lock also unlocks it, if locked.
	defer lock.Close()
	if err := lock.Lock(); err != nil {
		return err
	}
	// Perform the soft refresh viability check, possibly writing to the state
	// on failure.
	return inhibitRefresh(st, snapst, info, SoftNothingRunningRefreshCheck)
}
