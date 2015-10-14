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

package mutex

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

type node struct {
	sync.RWMutex
	count uint
}

type Map struct {
	mutex   sync.Mutex // who locks the locker? This guy.
	nodemap map[string]*node
	glock   *priv.Mutex
	gcount  uint
}

func New() *Map {
	return &Map{
		nodemap: make(map[string]*node),
		glock:   priv.New(dirs.SnapLockFile),
	}
}

func (lt *Map) lockGlobal() {
	if lt.gcount == 0 {
		if err := lt.glock.Lock(); err != nil {
			panic(err)
		}
	}
	lt.gcount++
}

func (lt *Map) unlockGlobal() {
	lt.gcount--
	if lt.gcount == 0 {
		if err := lt.glock.Unlock(); err != nil {
			panic(err)
		}
		// priv has an “interesting” api.
		lt.glock = priv.New(dirs.SnapLockFile)
	}
}

func (lt *Map) get(key string) *node {
	if _, ok := lt.nodemap[key]; !ok {
		lt.nodemap[key] = &node{}
	}

	return lt.nodemap[key]
}

func (lt *Map) rootNNode(key string) (root *node, node *node) {
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

func (lt *Map) Lock(args ...string) {
	if lt == nil {
		return
	}

	key := strings.Join(args, ".")
	root, node := lt.rootNNode(key)

	if key != "" {
		root.RLock()
	}

	node.Lock()
}

func (lt *Map) Unlock(args ...string) {
	if lt == nil {
		return
	}

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

func (lt *Map) RLock(args ...string) {
	if lt == nil {
		return
	}

	key := strings.Join(args, ".")
	root, node := lt.rootNNode(key)

	if key != "" {
		root.RLock()
	}

	node.RLock()
}

func (lt *Map) RUnlock(args ...string) {
	if lt == nil {
		return
	}

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
