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

package snapstate

import (
	"fmt"
	"time"

	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/overlord/configstate/config"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/strutil"
	"github.com/snapcore/snapd/timeutil"
)

// FIXME: what we actually want is a more flexible schedule spec that is
// user configurable  like:
// """
// tue
// tue,thu
// tue-thu
// 9:00
// 9:00,15:00
// 9:00-15:00
// tue,thu@9:00-15:00
// tue@9:00;thu@15:00
// mon,wed-fri@9:00-11:00,13:00-15:00
// """
// where 9:00 is implicitly taken as 9:00-10:00
// and tue is implicitly taken as tue@<our current setting?>
//
// it is controlled via:
// $ snap refresh --schedule=<time spec>
// which is a shorthand for
// $ snap set core refresh.schedule=<time spec>
// and we need to validate the time-spec, ideally internally by
// intercepting the set call

const defaultRefreshSchedule = "00:00-04:59/5:00-10:59/11:00-16:59/17:00-23:59"

type autoRefresh struct {
	state *state.State

	currentRefreshSchedule string
	nextRefresh            time.Time
	lastRefreshAttempt     time.Time
}

func newAutoRefresh(st *state.State) *autoRefresh {
	return &autoRefresh{state: st}
}

func (m *autoRefresh) Ensure() error {
	if ok, err := CanAutoRefresh(m.state); err != nil || !ok {
		return err
	}

	// get lastRefresh and schedule
	lastRefresh, err := m.LastRefresh()
	if err != nil {
		return err
	}
	refreshSchedule, err := m.checkRefreshSchedule()
	if err != nil {
		return err
	}
	if len(refreshSchedule) == 0 {
		return nil
	}

	// ensure nothing is in flight already
	if autoRefreshInFlight(m.state) {
		return nil
	}

	// compute next refresh attempt time (if needed)
	if m.nextRefresh.IsZero() {
		// store attempts in memory so that we can backoff
		if !lastRefresh.IsZero() {
			delta := timeutil.Next(refreshSchedule, lastRefresh)
			m.nextRefresh = time.Now().Add(delta)
		} else {
			// immediate
			m.nextRefresh = time.Now()
		}
		logger.Debugf("Next refresh scheduled for %s.", m.nextRefresh)
	}

	// Check that we have reasonable delays between unsuccessful attempts.
	// If the store is under stress we need to make sure we do not
	// hammer it too often
	if !m.lastRefreshAttempt.IsZero() && m.lastRefreshAttempt.Add(10*time.Minute).After(time.Now()) {
		return nil
	}

	// do refresh attempt (if needed)
	if !m.nextRefresh.After(time.Now()) {
		err = m.launchAutoRefresh()
		// clear nextRefresh only if the refresh worked. There is
		// still the lastRefreshAttempt rate limit so things will
		// not go into a busy store loop
		if err == nil {
			m.nextRefresh = time.Time{}
		}
	}

	return nil
}

func (m *autoRefresh) LastRefresh() (time.Time, error) {
	var lastRefresh time.Time
	err := m.state.Get("last-refresh", &lastRefresh)
	if err != nil && err != state.ErrNoState {
		return time.Time{}, err
	}
	return lastRefresh, nil
}

// NextRefresh returns the time the next update of the system's snaps
// will be attempted.
// The caller should be holding the state lock.
func (m *autoRefresh) NextRefresh() time.Time {
	return m.nextRefresh
}

// RefreshSchedule returns the current refresh schedule.
// The caller should be holding the state lock.
func (m *autoRefresh) RefreshSchedule() string {
	// This call ensures "m.currentRefreshSchedule" is up-to-date
	// with the latest configuration settings.
	m.checkRefreshSchedule()
	return m.currentRefreshSchedule
}

func autoRefreshInFlight(st *state.State) bool {
	for _, chg := range st.Changes() {
		if chg.Kind() == "auto-refresh" && !chg.Status().Ready() {
			return true
		}
	}
	return false
}

func (m *autoRefresh) launchAutoRefresh() error {
	m.lastRefreshAttempt = time.Now()
	updated, tasksets, err := AutoRefresh(m.state)
	if err != nil {
		logger.Noticef("Cannot prepare auto-refresh change: %s", err)
		return err
	}

	// Set last refresh time only if the store (in AutoRefresh) gave
	// us no error.
	m.state.Set("last-refresh", time.Now())

	var msg string
	switch len(updated) {
	case 0:
		logger.Noticef(i18n.G("No snaps to auto-refresh found"))
		return nil
	case 1:
		msg = fmt.Sprintf(i18n.G("Auto-refresh snap %q"), updated[0])
	case 2:
	case 3:
		quoted := strutil.Quoted(updated)
		// TRANSLATORS: the %s is a comma-separated list of quoted snap names
		msg = fmt.Sprintf(i18n.G("Auto-refresh snaps %s"), quoted)
	default:
		msg = fmt.Sprintf(i18n.G("Auto-refresh %d snaps"), len(updated))
	}

	chg := m.state.NewChange("auto-refresh", msg)
	for _, ts := range tasksets {
		chg.AddAll(ts)
	}
	chg.Set("snap-names", updated)
	chg.Set("api-data", map[string]interface{}{"snap-names": updated})

	return nil
}

var RefreshScheduleManaged func(st *state.State) bool

func refreshScheduleNoWeekdays(rs []*timeutil.Schedule) error {
	for _, s := range rs {
		if s.Weekday != "" {
			return fmt.Errorf("%q uses weekdays which is currently not supported", s)
		}
	}
	return nil
}

func (m *autoRefresh) checkRefreshSchedule() ([]*timeutil.Schedule, error) {
	if RefreshScheduleManaged != nil {
		if RefreshScheduleManaged(m.state) {
			if m.currentRefreshSchedule != "managed" {
				logger.Noticef("refresh.schedule is managed via the snapd-control interface")
				m.currentRefreshSchedule = "managed"
			}
			return nil, nil
		}
	}

	refreshScheduleStr := defaultRefreshSchedule
	tr := config.NewTransaction(m.state)
	err := tr.Get("core", "refresh.schedule", &refreshScheduleStr)
	if err != nil && !config.IsNoOption(err) {
		return nil, err
	}
	refreshSchedule, err := timeutil.ParseSchedule(refreshScheduleStr)
	if err == nil {
		err = refreshScheduleNoWeekdays(refreshSchedule)
	}
	if err != nil {
		logger.Noticef("cannot use refresh.schedule configuration: %s", err)
		refreshScheduleStr = defaultRefreshSchedule
		refreshSchedule, err = timeutil.ParseSchedule(refreshScheduleStr)
		if err != nil {
			panic(fmt.Sprintf("defaultRefreshSchedule cannot be parsed: %s", err))
		}
		tr.Set("core", "refresh.schedule", refreshScheduleStr)
		tr.Commit()
	}

	// we already have a refresh time, check if we got a new config
	if !m.nextRefresh.IsZero() {
		if m.currentRefreshSchedule != refreshScheduleStr {
			// the refresh schedule has changed
			logger.Debugf("Option refresh.schedule changed.")
			m.nextRefresh = time.Time{}
		}
	}
	m.currentRefreshSchedule = refreshScheduleStr

	return refreshSchedule, nil
}
