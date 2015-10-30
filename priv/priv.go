// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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

package priv

import (
	"errors"
	"syscall"
)

var (
	// ErrNeedRoot is return when an attempt to run a privileged operation
	// is made by an unprivileged process.
	ErrNeedRoot = errors.New("administrator privileges required")

	// ErrAlreadyLocked is returned when an attempts is made to lock an
	// already-locked FileLock.
	ErrAlreadyLocked = errors.New("another snappy is running, try again later")
)

// Mutex is the snappy mutual exclusion primitive.
type Mutex struct {
	lock *FileLock
}

// Determine if caller is running as the superuser
func isRootReal() bool {
	return syscall.Getuid() == 0
}

// useful for the tests
var isRoot = isRootReal

// New should be called when starting a privileged operation.
func New(fileName string) *Mutex {
	return &Mutex{
		lock: NewFileLock(fileName),
	}
}

// commonChecks encapsulates the checks that need to be run before any
// privileged operation.
func (m *Mutex) commonChecks() error {
	if !isRoot() {
		return ErrNeedRoot
	}

	return nil
}

// Lock attempts to acquire the mutex lock, and wil block if it is
// already locked.
func (m *Mutex) Lock() error {
	if err := m.commonChecks(); err != nil {
		return err
	}

	return m.lock.Lock(true)
}

// TryLock attempts to acquire the mutex lock. If it is already locked,
// it will return ErrAlreadyLocked.
func (m *Mutex) TryLock() error {
	if err := m.commonChecks(); err != nil {
		return err
	}

	return m.lock.Lock(false)
}

// Unlock will unlock the specified mutex, returning ErrNotLocked if
// the mutex is not already locked.
func (m *Mutex) Unlock() error {
	if err := m.commonChecks(); err != nil {
		return err
	}

	if m.lock == nil {
		return ErrNotLocked
	}

	err := m.lock.Unlock()
	if err != nil {
		return err
	}

	return nil
}

// WithMutex runs the function f with the priv.Mutex hold
func WithMutex(fileName string, f func() error) error {
	privMutex := New(fileName)
	if err := privMutex.TryLock(); err != nil {
		return err
	}
	defer privMutex.Unlock()

	return f()
}
