// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
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
	"strings"

	"github.com/snapcore/snapd/boot"
	"github.com/snapcore/snapd/overlord/snapstate"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/secboot"
)

// use as:
// sudo curl -sS --unix-socket /run/snapd.socket http://localhost/v2/debug -X POST -d '{"action":"fde-setup","message":"mock-kernel-snap,key,key-name"}'
func debugFdeSetup(st *state.State, msg string) Response {
	l := strings.Split(msg, ",")
	if len(l) != 3 {
		return BadRequest("msg should be three comma separated stings, ot: %q", msg)
	}
	mockOp := l[0]
	mockKernelSnapName := l[1]
	mockKey := l[2]
	mockKeyName := l[3]

	info, err := snapstate.CurrentInfo(st, mockKernelSnapName)
	if err != nil {
		return BadRequest("cannot get info for %s: %v", mockKernelSnapName, err)
	}
	var mk secboot.EncryptionKey
	copy(mk[:], []byte(mockKey))
	params := &boot.FdeSetupHookParams{
		KernelInfo: info,
		Key:        mk,
		KeyName:    mockKeyName,
	}

	sealedKey, err := boot.FdeSetupHookRunner(mockOp, params)
	if err != nil {
		return BadRequest("hook failed with: %v", err)
	}

	return SyncResponse(string(sealedKey), nil)
}
