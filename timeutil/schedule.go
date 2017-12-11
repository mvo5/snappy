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

package timeutil

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Match 0:00-24:00, where 24:00 means the later end of the day.
var validTime = regexp.MustCompile(`^([0-9]|0[0-9]|1[0-9]|2[0-3]):([0-5][0-9])$|^24:00$`)

// Clock represents a hour:minute time within a day.
type Clock struct {
	Hour   int
	Minute int
}

func (t Clock) String() string {
	return fmt.Sprintf("%02d:%02d", t.Hour, t.Minute)
}

// Sub returns the duration t - other.
func (t Clock) Sub(other Clock) time.Duration {
	t1 := time.Duration(t.Hour)*time.Hour + time.Duration(t.Minute)*time.Minute
	t2 := time.Duration(other.Hour)*time.Hour + time.Duration(other.Minute)*time.Minute
	return t1 - t2
}

// Add adds given duration to t and returns a new Clock
func (t Clock) Add(dur time.Duration) Clock {
	t1 := time.Duration(t.Hour)*time.Hour + time.Duration(t.Minute)*time.Minute
	t2 := t1 + dur
	nt := Clock{
		Hour:   int(t2.Hours()) % 24,
		Minute: int(t2.Minutes()) % 60,
	}
	return nt
}

// Time generates a time.Time with hour and minute set from t, while year, month
// and day are taken from base
func (t Clock) Time(base time.Time) time.Time {
	return time.Date(base.Year(), base.Month(), base.Day(),
		t.Hour, t.Minute, 0, 0, base.Location())
}

// IsValidWeekday returns true if given s looks like a valid weekday. Valid
// inputs are 3 letter, lowercase abbreviations of week days.
func IsValidWeekday(s string) bool {
	_, ok := weekdayMap[s]
	return ok
}

// ParseClock parses a string that contains hour:minute and returns
// an Clock type or an error
func ParseClock(s string) (t Clock, err error) {
	m := validTime.FindStringSubmatch(s)
	if len(m) == 0 {
		return t, fmt.Errorf("cannot parse %q", s)
	}

	if m[0] == "24:00" {
		t.Hour = 24
		return t, nil
	}

	t.Hour, err = strconv.Atoi(m[1])
	if err != nil {
		return t, fmt.Errorf("cannot parse %q: %s", m[1], err)
	}
	t.Minute, err = strconv.Atoi(m[2])
	if err != nil {
		return t, fmt.Errorf("cannot parse %q: %s", m[2], err)
	}
	return t, nil
}

// Week represents a weekday such as Monday, Tuesday, with optional
// week-in-the-month position, eg. the first Monday of the month
type Week struct {
	Weekday time.Weekday
	// Pos defines which week inside the month the Day refers to, where zero
	// means every week, 1 means first occurrence of the weekday, and 5
	// means last occurrence (which might be the fourth or the fifth).
	Pos uint
}

func (w Week) String() string {
	// Wednesday -> wed
	day := strings.ToLower(w.Weekday.String()[0:3])
	if w.Pos == 0 {
		return day
	}
	return day + strconv.Itoa(int(w.Pos))
}

// WeekSpan represents a span of weekdays between Start and End days. WeekSpan
// may wrap around the week, eg. fri-mon is a span from Friday to Monday
type WeekSpan struct {
	Start Week
	End   Week
}

func (ws WeekSpan) String() string {
	if ws.End != ws.Start {
		return ws.Start.String() + "-" + ws.End.String()
	}
	return ws.Start.String()
}

// Match checks if t is within the day span represented by ws.
func (ws WeekSpan) Match(t time.Time) bool {
	start, end := ws.Start, ws.End
	wdStart, wdEnd := start.Weekday, end.Weekday

	// is it the right week?
	if start.Pos > 0 {
		week := uint(t.Day()/7) + 1

		if start.Pos == 5 {
			// last week of the month
			if !isLastWeekdayInMonth(t) {
				return false
			}
		} else {
			if week < start.Pos || week > end.Pos {
				return false
			}
		}
	}

	if wdStart <= wdEnd {
		// single day (mon) or start < end (eg. mon-fri)
		return t.Weekday() >= wdStart && t.Weekday() <= wdEnd
	}
	// wraps around the week end, eg. fri-mon
	return t.Weekday() >= wdStart || t.Weekday() <= wdEnd
}

// ClockSpan represents a time span within 24h, potentially crossing days. For
// example, 23:00-1:00 represents a span from 11pm to 1am.
type ClockSpan struct {
	Start Clock
	End   Clock
	// Split defines the number of subspans this span will be divided into.
	Split uint
	// Spread defines whether the events are randomly spread inside the span
	// or subspans.
	Spread bool
}

func (ts ClockSpan) String() string {
	sep := "-"
	if ts.Spread {
		sep = "~"
	}
	if ts.End != ts.Start {
		s := ts.Start.String() + sep + ts.End.String()
		if ts.Split > 0 {
			s += "/" + strconv.Itoa(int(ts.Split))
		}
		return s
	}
	return ts.Start.String()
}

// Window generates a ScheduleWindow which has the start date same as t. The
// window's start and end time are set according to Start and End, with the end
// time possibly crossing into the next day.
func (ts ClockSpan) Window(t time.Time) ScheduleWindow {
	start := ts.Start.Time(t)
	end := ts.End.Time(t)

	// 23:00-1:00
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
	}
	return ScheduleWindow{
		Start:  start,
		End:    end,
		Spread: ts.Spread,
	}
}

// Subspans returns a slice of TimeSpan generated from ts by splitting the time
// between ts.Start and ts.End into ts.Split equal spans.
func (ts ClockSpan) Subspans() []ClockSpan {
	if ts.Split == 0 || ts.Split == 1 || ts.End == ts.Start {
		return []ClockSpan{ts}
	}

	span := ts.End.Sub(ts.Start)
	step := span / time.Duration(ts.Split)

	spans := make([]ClockSpan, ts.Split)
	for i := uint(0); i < ts.Split; i++ {
		start := ts.Start.Add(time.Duration(i) * step)
		spans[i] = ClockSpan{
			Start:  start,
			End:    start.Add(step),
			Split:  0, // no more subspans
			Spread: ts.Spread,
		}
	}
	return spans
}

// Schedule represents a single schedule
type Schedule struct {
	WeekSpans  []WeekSpan
	ClockSpans []ClockSpan
}

func (sched *Schedule) String() string {
	buf := &bytes.Buffer{}

	for i, span := range sched.WeekSpans {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(span.String())
	}

	if len(sched.WeekSpans) > 0 && len(sched.ClockSpans) > 0 {
		buf.WriteByte(',')
	}

	for i, span := range sched.ClockSpans {
		if i > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(span.String())
	}
	return buf.String()
}

func (sched *Schedule) weekSpans() []WeekSpan {
	return sched.WeekSpans
}

func (sched *Schedule) flattenedTimeSpans() []ClockSpan {
	baseTimes := sched.ClockSpans
	if len(baseTimes) == 0 {
		baseTimes = []ClockSpan{{}}
	}

	times := make([]ClockSpan, 0, len(baseTimes))
	for _, ts := range baseTimes {
		times = append(times, ts.Subspans()...)
	}
	return times
}

// isLastWeekdayInMonth returns true if t.Weekday() is the last weekday
// occurring this t.Month(), eg. check is Feb 25 2017 is the last Saturday of
// February.
func isLastWeekdayInMonth(t time.Time) bool {
	// try a week from now, if it's still the same month then t.Weekday() is
	// not last
	return t.Month() != t.Add(7*24*time.Hour).Month()
}

// ScheduleWindow represents a time window between Start and End times when the
// scheduled event can happen.
type ScheduleWindow struct {
	Start time.Time
	End   time.Time
	// Spread defines whether the event shall be randomly placed between
	// Start and End times
	Spread bool
}

// StartsBefore returns true if the window's start time is before t.
func (s ScheduleWindow) StartsBefore(t time.Time) bool {
	return s.Start.Before(t)
}

// EndsBefore returns true if the window's end time is before t.
func (s ScheduleWindow) EndsBefore(t time.Time) bool {
	return s.End.Before(t)
}

// Returnes true if t falls inside the window.
func (s ScheduleWindow) Includes(t time.Time) bool {
	return (t.Equal(s.Start) || t.After(s.Start)) && (t.Equal(s.End) || t.Before(s.End))
}

// Returns true if window is uninitialized.
func (s ScheduleWindow) IsZero() bool {
	return s.Start.IsZero()
}

// Next returns when the time of the next interval defined in sched.
func (sched *Schedule) Next(last time.Time) ScheduleWindow {
	now := timeNow()

	wspans := sched.weekSpans()
	tspans := sched.flattenedTimeSpans()

	for t := last; ; t = t.Add(24 * time.Hour) {
		// try to find a matching schedule by moving in 24h jumps, check
		// if the event needs to happen on a specific day in a specific
		// week, next pick the earliest event time

		var window ScheduleWindow

		// if there's a week schedule, check if we hit that first
		if len(wspans) > 0 {
			var weekMatch bool
			for _, week := range wspans {
				if week.Match(t) {
					weekMatch = true
					break
				}
			}

			if !weekMatch {
				continue
			}
		}

		for _, tspan := range tspans {
			// consider all time spans for this particular date and
			// find the earliest possible one that is not before
			// 'now', and does not include the 'last' time
			newWindow := tspan.Window(t)

			// the time span ends before 'now', try another one
			if newWindow.EndsBefore(now) {
				continue
			}

			// same interval as last update, move forward
			if newWindow.Includes(last) {
				continue
			}

			// if this candidate comes before current candidate use it
			if window.IsZero() || newWindow.StartsBefore(window.Start) {
				window = newWindow
			}
		}
		// no suitable time span was found this day so try the next day
		if window.EndsBefore(now) {
			continue
		}
		return window
	}

}

func randDur(a, b time.Time) time.Duration {
	dur := b.Sub(a)
	if dur > 5*time.Minute {
		// doing it this way we still spread really small windows about
		dur -= 5 * time.Minute
	}

	if dur <= 0 {
		// avoid panic'ing (even if things are probably messed up)
		return 0
	}

	return time.Duration(rand.Int63n(int64(dur)))
}

var (
	timeNow = time.Now

	// FIMXE: pass in as a parameter for next
	maxDuration = 60 * 24 * time.Hour
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Next returns the earliest event after last according to the provided
// schedule.
func Next(schedule []*Schedule, last time.Time) time.Duration {
	now := timeNow()

	window := ScheduleWindow{
		Start: last.Add(maxDuration),
		End:   last.Add(maxDuration).Add(1 * time.Hour),
	}

	for _, sched := range schedule {
		next := sched.Next(last)
		if next.StartsBefore(window.Start) {
			window = next
		}
	}
	if window.StartsBefore(now) {
		return 0
	}

	when := window.Start.Sub(now)
	if window.Spread {
		when += randDur(window.Start, window.End)
	}

	return when

}

var weekdayMap = map[string]time.Weekday{
	"sun": time.Sunday,
	"mon": time.Monday,
	"tue": time.Tuesday,
	"wed": time.Wednesday,
	"thu": time.Thursday,
	"fri": time.Friday,
	"sat": time.Saturday,
}

// parseTimeInterval gets an input like "9:00-11:00"
// and extracts the start and end of that schedule string and
// returns them and any errors.
func parseTimeInterval(s string) (start, end Clock, err error) {
	l := strings.SplitN(s, "-", 2)
	if len(l) != 2 {
		return start, end, fmt.Errorf("cannot parse %q: not a valid interval", s)
	}

	start, err = ParseClock(l[0])
	if err != nil {
		return start, end, fmt.Errorf("cannot parse %q: not a valid time", l[0])
	}
	end, err = ParseClock(l[1])
	if err != nil {
		return start, end, fmt.Errorf("cannot parse %q: not a valid time", l[1])
	}

	return start, end, nil
}

// parseSingleSchedule parses a schedule string like "9:00-11:00"
func parseSingleSchedule(s string) (*Schedule, error) {
	start, end, err := parseTimeInterval(s)
	if err != nil {
		return nil, err
	}

	return &Schedule{
		ClockSpans: []ClockSpan{
			{
				Start:  start,
				End:    end,
				Spread: true,
			},
		},
	}, nil
}

// ParseLegacySchedule takes a schedule string in the form of:
//
// 9:00-15:00 (every day between 9am and 3pm)
// 9:00-15:00/21:00-22:00 (every day between 9am,5pm and 9pm,10pm)
//
// and returns a list of Schedule types or an error
func ParseLegacySchedule(scheduleSpec string) ([]*Schedule, error) {
	var schedule []*Schedule

	for _, s := range strings.Split(scheduleSpec, "/") {
		sched, err := parseSingleSchedule(s)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, sched)
	}

	return schedule, nil
}

// ParseSchedule parses a schedule in V2 format. The format is described as:
//
//     eventlist = eventset *( ".." eventset )
//     eventset = wdaylist / timelist / wdaylist "," timelist
//
//     wdaylist = wdayset *( "," wdayset )
//     wdayset = wday / wdayspan
//     wday =  ( "sun" / "mon" / "tue" / "wed" / "thu" / "fri" / "sat" ) [ DIGIT ]
//     wdayspan = wday "-" wday
//
//     timelist = timeset *( "," timeset )
//     timeset = time / timespan
//     time = 2DIGIT ":" 2DIGIT
//     timespan = time ( "-" / "~" ) time [ "/" ( time / count ) ]
//     count = 1*DIGIT
//
// Examples:
// mon,10:00..fri,15:00 (Monday at 10:00, Friday at 15:00)
// mon,fri,10:00,15:00 (Monday at 10:00 and 15:00, Friday at 10:00 and 15:00)
// mon-wed,fri,9:00-11:00/2 (Monday to Wednesday and on Friday, twice between
//                           9:00 and 11:00)
// mon,9:00~11:00..wed,22:00~23:00 (Monday, sometime between 9:00 and 11:00, and
//                                  on Wednesday, sometime between 22:00 and 23:00)
// mon,wed  (Monday and on Wednesday)
// mon..wed (same as above)
//
// Returns a slice of schedules or an error if parsing failed
func ParseSchedule(scheduleSpec string) ([]*Schedule, error) {
	var schedule []*Schedule

	for _, s := range strings.Split(scheduleSpec, "..") {
		// cut the schedule in event sets
		//     eventlist = eventset *( ".." eventset )
		sched, err := parseEventSet(s)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, sched)
	}
	return schedule, nil
}

// parseWeekSpan generates a per day (span) schedule
func parseWeekSpan(start string, end string) (*WeekSpan, error) {
	startWeekday, err := parseWeekday(start)
	if err != nil {
		return nil, err
	}

	endWeekday, err := parseWeekday(end)
	if err != nil {
		return nil, err
	}

	if endWeekday.Pos < startWeekday.Pos {
		return nil, fmt.Errorf("unsupported schedule")
	}

	if (startWeekday.Pos != 0 && endWeekday.Pos == 0) ||
		(endWeekday.Pos != 0 && startWeekday.Pos == 0) {
		return nil, fmt.Errorf("mixed weekday and nonweekday")
	}

	span := &WeekSpan{
		Start: *startWeekday,
		End:   *endWeekday,
	}

	return span, nil
}

// parseSpan splits a span string `<start>-<end>` and returns the start and end
// pieces. If string is not a span then the whole string is returned
func parseSpan(spec string) (string, string) {
	var start, end string

	specs := strings.Split(spec, spanToken)

	if len(specs) == 1 {
		start = specs[0]
		end = start
	} else {
		start = specs[0]
		end = specs[1]
	}
	return start, end
}

// parseTime parses a time specification which can either be `<hh>:<mm>` or
// `<hh>:<mm>-<hh>:<mm>`. Returns corresponding Clock structs.
func parseTime(start, end string) (Clock, Clock, error) {
	// is it a time?
	var err error
	var tstart, tend Clock

	if start == end {
		// single time, eg. 10:00
		tstart, err = ParseClock(start)
		if err == nil {
			tend = tstart
		}
	} else {
		// time span, eg. 10:00-12:00
		tstart, tend, err = parseTimeInterval(fmt.Sprintf("%s-%s", start, end))
	}

	return tstart, tend, err
}

func parseTimeSpan(start, end string) (*ClockSpan, error) {

	startTime, endTime, err := parseTime(start, end)
	if err != nil {
		return nil, err
	}

	span := &ClockSpan{
		Start: startTime,
		End:   endTime,
	}
	return span, nil
}

// parseWeekday will parse a string and extract weekday (eg. wed),
// MonthWeekday (eg. 1st, 5th in the month).
func parseWeekday(s string) (*Week, error) {
	l := len(s)
	if l != 3 && l != 4 {
		return nil, fmt.Errorf("cannot parse %q: invalid format", s)
	}

	day := s
	var pos uint
	if l == 4 {
		day = s[0:3]
		if v, err := strconv.ParseUint(s[3:], 10, 32); err != nil {
			return nil, fmt.Errorf("cannot parse %q: invalid week", s)
		} else if v < 1 || v > 5 {
			return nil, fmt.Errorf("cannot parse %q: incorrect week number", s)
		} else {
			pos = uint(v)
		}
	}

	if !IsValidWeekday(day) {
		return nil, fmt.Errorf("cannot parse %q: invalid weekday", s)
	}

	week := &Week{
		Weekday: weekdayMap[day],
		Pos:     pos,
	}
	return week, nil
}

// parseCount will parse the string containing a count token and return the
// count, the rest of the string with count information removed or an error
func parseCount(s string) (count uint, rest string, err error) {
	if strings.Contains(s, countToken) {
		//     timespan = time ( "-" / "~" ) time [ "/" ( time / count ) ]
		ws := strings.Split(s, countToken)
		rest = ws[0]
		countStr := ws[1]
		c, err := strconv.ParseUint(countStr, 10, 32)
		if err != nil || c == 0 {
			return 0, "", fmt.Errorf("cannot parse %q: not a valid event interval",
				s)
		}
		return uint(c), rest, nil
	}
	return 0, s, nil
}

// isValidWeekdaySpec returns true if string is of the following format:
//     wday =  ( "sun" / "mon" / "tue" / "wed" / "thu" / "fri" / "sat" ) [ DIGIT ]
func isValidWeekdaySpec(s string) bool {
	_, err := parseWeekday(s)
	return err == nil
}

const (
	spanToken           = "-"
	randomizedSpanToken = "~"
	countToken          = "/"
)

var specialTokens = map[string]string{
	// shorthand variant of whole day
	"-": "0:00-24:00",
	// and randomized whole day
	"~": "0:00~24:00",
}

// Parse each event and return a slice of schedules.
func parseEventSet(s string) (*Schedule, error) {
	var events []string
	// split event set into events
	//     eventset = wdaylist / timelist / wdaylist "," timelist
	// or wdaysets
	//     wdaylist = wdayset *( "," wdayset )
	// or timesets
	//     timelist = timeset *( "," timeset )
	//
	// NOTE: the syntax is ambiguous in the sense the type of a 'set' is now
	// explicitly indicated, thus the parsing will first try to handle it as
	// a wdayset, then as timeset

	if els := strings.Split(s, ","); len(els) > 1 {
		events = els
	} else {
		events = []string{s}
	}

	var schedule Schedule
	// indicates that any further events must be time events
	var expectTime bool

	for _, event := range events {
		// process events one by one

		randomized := false
		var count uint
		var start string
		var end string

		if c, rest, err := parseCount(event); err != nil {
			return nil, err
		} else if c > 0 {
			// count was specified
			count = c
			// count token is only allowed in timespans, meaning
			// we're parsing only timespans from now on
			expectTime = true

			// update the remaining part of the event
			event = rest
		}

		if special, ok := specialTokens[event]; ok {
			event = special
		}

		if strings.Contains(event, randomizedSpanToken) {
			// timespan uses "~" to indicate that the actual event
			// time is to be randomized.
			randomized = true
			event = strings.Replace(event,
				randomizedSpanToken,
				spanToken, 1)

			// randomize token is only allowed in timespans, meaning
			// we're parsing only timespans from now on
			expectTime = true
		}

		// spans
		//     wdayspan = wday "-" wday
		//     timespan = time ( "-" / "~" ) time [ "/" ( time / count ) ]
		start, end = parseSpan(event)

		if validTime.MatchString(start) && validTime.MatchString(end) {
			// from now on we expect only timespans
			expectTime = true

			if span, err := parseTimeSpan(start, end); err != nil {
				return nil, err
			} else {
				span.Split = count
				span.Spread = randomized

				schedule.ClockSpans = append(schedule.ClockSpans, *span)
			}

		} else if isValidWeekdaySpec(start) && isValidWeekdaySpec(end) {
			// is it a day?

			if expectTime {
				return nil, fmt.Errorf("cannot parse %q: expected time spec", s)
			}

			if span, err := parseWeekSpan(start, end); err != nil {
				return nil, fmt.Errorf("cannot parse %q: %s", event, err.Error())
			} else {
				schedule.WeekSpans = append(schedule.WeekSpans, *span)
			}
		} else {
			// no, it's an error
			return nil, fmt.Errorf("cannot parse %q: not a valid event", s)
		}

	}

	return &schedule, nil
}
