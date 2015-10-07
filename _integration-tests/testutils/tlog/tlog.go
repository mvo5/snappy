// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015 Canonical Ltd
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

package tlog

import (
	"fmt"
	"io"
	"os"
)

// Level represents the logging level
type Level uint8

// ErrLogLevel is returned when a not recognized textual log level is passed
// to SetTextLevel
type ErrLogLevel struct {
	s string
}

func (e *ErrLogLevel) Error() string {
	return fmt.Sprintf("The level %s is not supported", e.s)
}

const (
	// InfoLevel is the less verbose mode
	InfoLevel Level = iota
	// DebugLevel is the most verbose mode
	DebugLevel
)

// Logger is the type of the internal instance holding the state
type Logger struct {
	output io.Writer // the output goes through this writer
	level  Level     // log level, functions with a level below this won't write to output
}

var l *Logger

func init() {
	l = &Logger{output: os.Stdout, level: DebugLevel}
}

// GetOutput is the getter for the output writter
func GetOutput() io.Writer {
	return l.output
}

// SetOutput is the setter for the output writter
func SetOutput(w io.Writer) {
	l.output = w
}

// GetLevel is the getter for the log level
func GetLevel() Level {
	return l.level
}

// SetLevel is the getter for the log level
func SetLevel(lvl Level) {
	l.level = lvl
}

// SetTextLevel is the getter for the log level
func SetTextLevel(s string) (err error) {
	switch s {
	case "info":
		SetLevel(InfoLevel)
		return
	case "debug":
		SetLevel(DebugLevel)
		return
	default:
		return &ErrLogLevel{s}
	}
}

// Debugf outputs debug messages
func Debugf(format string, args ...interface{}) {
	if l.level >= DebugLevel {
		fmt.Fprintf(l.output, format, args...)
	}
}

// Infof outputs info messages
func Infof(format string, args ...interface{}) {
	fmt.Fprintf(l.output, format, args...)
}
