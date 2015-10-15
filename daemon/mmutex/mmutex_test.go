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

package mmutex

import (
	"errors"
	"sync"
	"testing"
	"time"

	"gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type mmutexSuite struct {
	oldGlock func() privLocker
	l        *muLck
}

var _ = check.Suite(&mmutexSuite{})

type muLck struct {
	m sync.Mutex
	e error
}

func (mu *muLck) Lock() error   { mu.m.Lock(); return mu.e }
func (mu *muLck) Unlock() error { mu.m.Unlock(); return mu.e }

func (s *mmutexSuite) SetUpTest(c *check.C) {
	s.oldGlock = newGlock
	s.l = &muLck{}
	newGlock = func() privLocker {
		return s.l
	}
}

func (s *mmutexSuite) TearDownTest(c *check.C) {
	newGlock = s.oldGlock
}

func (s *mmutexSuite) TestSimultaneousLocks(c *check.C) {
	ch := make(chan bool)

	mm := New().(*mmutex)
	go func() {
		mm.Lock("foo")
		defer mm.Unlock("foo")
		mm.Lock("bar")
		defer mm.Unlock("bar")
		c.Check(mm.nodemap, check.HasLen, 3)
		c.Check(mm.nodemap[""], check.NotNil)
		c.Check(mm.nodemap["foo"], check.NotNil)
		c.Check(mm.nodemap["bar"], check.NotNil)

		ch <- true
	}()
	select {
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	case <-ch:
	}

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestSimultaneousRLocks(c *check.C) {
	ch := make(chan bool)

	mm := New().(*mmutex)
	go func() {
		mm.RLock("foo")
		defer mm.RUnlock("foo")
		mm.RLock("bar")
		defer mm.RUnlock("bar")
		c.Check(mm.nodemap, check.HasLen, 3)
		c.Check(mm.nodemap[""], check.NotNil)
		c.Check(mm.nodemap["foo"], check.NotNil)
		c.Check(mm.nodemap["bar"], check.NotNil)

		ch <- true
	}()
	select {
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	case <-ch:
	}

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestSimultaneousRLockWRoot(c *check.C) {
	ch := make(chan bool)

	mm := New().(*mmutex)
	go func() {
		mm.RLock()
		defer mm.RUnlock()
		mm.RLock("foo")
		defer mm.RUnlock("foo")
		c.Check(mm.nodemap, check.HasLen, 2)
		c.Check(mm.nodemap[""], check.NotNil)
		c.Check(mm.nodemap["foo"], check.NotNil)

		ch <- true
	}()
	select {
	case <-time.After(time.Second):
		c.Fatalf("timed out")
	case <-ch:
	}

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestConcurrentRootLockWRLock(c *check.C) {
	ch := make(chan bool)
	buf := make(chan bool, 10)

	mm := New().(*mmutex)
	go func() {
		mm.RLock("foo")
		<-ch
		defer mm.RUnlock("foo")

		time.Sleep(100 * time.Millisecond)

		buf <- false
	}()

	ch <- true // waits for mm.RLock("foo")

	go func() {
		mm.Lock()
		defer mm.Unlock()

		buf <- true
	}()

	c.Check(<-buf, check.Equals, false)
	c.Check(<-buf, check.Equals, true)

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestConcurrentRLockWRootLock(c *check.C) {
	ch := make(chan bool)
	buf := make(chan bool, 10)

	mm := New().(*mmutex)
	go func() {
		mm.Lock()
		defer mm.Unlock()
		<-ch

		time.Sleep(100 * time.Millisecond)

		buf <- true
	}()

	ch <- true // waits for mm.Lock()

	go func() {
		mm.RLock("foo")
		defer mm.RUnlock("foo")

		buf <- false
	}()

	c.Check(<-buf, check.Equals, true)
	c.Check(<-buf, check.Equals, false)

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestConcurrentRootLockWLock(c *check.C) {
	ch := make(chan bool)
	buf := make(chan bool, 10)

	mm := New().(*mmutex)
	go func() {
		mm.Lock("foo")
		<-ch
		defer mm.Unlock("foo")

		time.Sleep(100 * time.Millisecond)

		buf <- false
	}()

	ch <- true // waits for mm.Lock("foo")

	go func() {
		mm.Lock()
		defer mm.Unlock()

		buf <- true
	}()

	c.Check(<-buf, check.Equals, false)
	c.Check(<-buf, check.Equals, true)

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestConcurrentLockWRootLock(c *check.C) {
	ch := make(chan bool)
	buf := make(chan bool, 10)

	mm := New().(*mmutex)
	go func() {
		mm.Lock()
		<-ch
		defer mm.Unlock()

		time.Sleep(100 * time.Millisecond)

		buf <- true
	}()

	ch <- true // waits for mm.Lock()

	go func() {
		mm.Lock("foo")
		defer mm.Unlock("foo")

		buf <- false
	}()

	c.Check(<-buf, check.Equals, true)
	c.Check(<-buf, check.Equals, false)

	c.Check(mm.nodemap, check.HasLen, 1)
	c.Check(mm.nodemap[""], check.NotNil)
}

func (s *mmutexSuite) TestLockPanicsIfGlobalLockFails(c *check.C) {
	s.l.e = errors.New("failed")
	mm := New().(*mmutex)
	c.Check(func() { mm.Lock() }, check.Panics, s.l.e)
}

func (s *mmutexSuite) TestUnlockPanicsIfGlobalUnlockFails(c *check.C) {
	mm := New().(*mmutex)
	mm.Lock()
	s.l.e = errors.New("failed")

	c.Check(func() { mm.Unlock() }, check.Panics, s.l.e)
}
