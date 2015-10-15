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
	"strings"
	"sync"

	"launchpad.net/snappy/dirs"
	"launchpad.net/snappy/priv"
)

/*

(R)Lock:

1. lock self
2. lock global:
  a. globalCount == 0? actually lock global
  b. increase globalCount
3. get locks for "" and key
4. unlock self
5. key != "" ? -> RLock ""
6. (R)Lock key

(R)Unlock:

1. lock self
2. (R)Unlock key
3. key != "" ? -> RUnlock ""
4. unlock global:
  a. decrease globalCount
  b. 0 globalCount? actually unlock global
5. unlock self

*/

// privLocker is the interface of priv.Mutex that we use
type privLocker interface {
	Lock() error
	Unlock() error
}

// newGlock calls priv.New; override it in tests to not panic when not
// root.
var newGlock = func() privLocker {
	return priv.New(dirs.SnapLockFile)
}

type node struct {
	sync.RWMutex
	count uint
}

// A MMutex is a map of mutexes, with a special "root" mutex that lets you
// Lock or RLock any of the non-root mutexes, but if the root mutex is
// Locked you'll have to wait for it.
type MMutex interface {
	Lock(...string)
	RLock(...string)
	Unlock(...string)
	RUnlock(...string)
}

type mmutex struct {
	mutex   sync.Mutex // who locks the locker? This guy.
	nodemap map[string]*node
	glock   privLocker
	gcount  uint
}

// New is the mmutex constructor.
func New() MMutex {
	return &mmutex{
		nodemap: make(map[string]*node),
	}
}

// acquires/increases count of the global lock
// global lock only needed to interop with non-rest-api snappy
func (lt *mmutex) lockGlobal() {
	if lt.gcount == 0 {
		lt.glock = newGlock()
		if err := lt.glock.Lock(); err != nil {
			panic(err)
		}
	}
	lt.gcount++
}

// decreases count and releases the global lock
// global lock only needed to interop with non-rest-api snappy
func (lt *mmutex) unlockGlobal() {
	lt.gcount--
	if lt.gcount == 0 {
		if err := lt.glock.Unlock(); err != nil {
			panic(err)
		}
		// priv's Unlock sets the internal, private lock to nil;
		// might as well get rid of the lock ourselves as it's
		// useless at this point.
		lt.glock = nil
	}
}

// convenience autovivifying getter
func (lt *mmutex) get(key string) *node {
	if _, ok := lt.nodemap[key]; !ok {
		lt.nodemap[key] = &node{}
	}

	return lt.nodemap[key]
}

// get the "root" lock (lock for "") and the lock for key
func (lt *mmutex) rootNNode(key string) (root *node, node *node) {
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	lt.lockGlobal()

	root = lt.get("")
	node = lt.get(key)

	if key != "" {
		node.count++
	}

	return root, node
}

// Lock the mutex for the given key. If the optional args are given, the
// root mutex is RLocked; otherwise, the root mutex itself is Locked.
func (lt *mmutex) Lock(args ...string) {
	key := strings.Join(args, ".")
	root, node := lt.rootNNode(key)

	if key != "" {
		root.RLock()
	}

	node.Lock()
}

// Unlock the mutex for the given key, and RUnlock the root mutex if args
// are given.
func (lt *mmutex) Unlock(args ...string) {
	key := strings.Join(args, ".")
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	node := lt.get(key)
	node.Unlock()
	if key != "" {
		lt.get("").RUnlock()
		node.count--
		if node.count == 0 {
			delete(lt.nodemap, key)
		}
	}

	lt.unlockGlobal()
}

// RLock the mutex for the given key; also RLock the root mutex if args are
// given.
func (lt *mmutex) RLock(args ...string) {
	key := strings.Join(args, ".")
	root, node := lt.rootNNode(key)

	if key != "" {
		root.RLock()
	}

	node.RLock()
}

// RUnlock the mutex for the given key; also RUnlock the root mutex if args
// are given.
func (lt *mmutex) RUnlock(args ...string) {
	key := strings.Join(args, ".")
	lt.mutex.Lock()
	defer lt.mutex.Unlock()

	node := lt.get(key)
	node.RUnlock()
	if key != "" {
		lt.get("").RUnlock()
		node.count--
		if node.count == 0 {
			delete(lt.nodemap, key)
		}
	}

	lt.unlockGlobal()
}
