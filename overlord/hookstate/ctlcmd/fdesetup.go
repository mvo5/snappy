// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020q Canonical Ltd
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

package ctlcmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/secboot"
)

type fdeSetupRequestCommand struct {
	baseCommand
}

var shortFdeSetupRequestHelp = i18n.G("Request setup of full disk encryption")
var longFdeSetupRequestHelp = i18n.G(`
The fde-setup-request command returns the data to setup full disk encryption.

    $ snapctl fde-setup-request
{"key": "some-key"}
`)

func init() {
	addCommand("fde-setup-request", shortFdeSetupRequestHelp, longFdeSetupRequestHelp, func() command { return &fdeSetupRequestCommand{} })
}

type fdeSetupJSON struct {
	FdeKey       secboot.EncryptionKey `json:"fde-key,omitempty"`
	FdeSealedKey secboot.EncryptionKey `json:"fde-sealed-key,omitempty"`
}

func (c *fdeSetupRequestCommand) Execute(args []string) error {
	context := c.context()
	if context == nil {
		return fmt.Errorf("cannot  without a context")
	}

	var js fdeSetupJSON
	if err := context.Get("fde-key", &js.FdeKey); err != nil {
		return fmt.Errorf("cannot get fde key from context: %v", err)
	}
	bytes, err := json.MarshalIndent(js, "", "\t")
	if err != nil {
		return fmt.Errorf("cannot json print fde key: %v", err)
	}
	c.printf("%s\n", string(bytes))

	return nil
}

type fdeSetupResultCommand struct {
	baseCommand
}

var shortFdeSetupResultHelp = i18n.G("Set result for FDE key sealing")
var longFdeSetupResultHelp = i18n.G(`
The fde-setup-result command reads the result data from a FDE setu.

    $ echo "{"fde-sealed-key":"the-sealed-key"} | snapctl fde-setup-result
`)

func init() {
	addCommand("fde-setup-result", shortFdeSetupResultHelp, longFdeSetupResultHelp, func() command { return &fdeSetupResultCommand{} })
}

var Stdin = os.Stdin

func (c *fdeSetupResultCommand) Execute(args []string) error {
	context := c.context()
	if context == nil {
		return fmt.Errorf("cannot  without a context")
	}

	// XXX: should we do something simpler than json?
	var js fdeSetupJSON
	// XXX: will reading from stdin work in hooks?
	if err := json.NewDecoder(Stdin).Decode(&js); err != nil {
		return fmt.Errorf("cannot parse fde-setup-result json: %v", err)
	}
	context.Set("fde-sealed-key", &js.FdeSealedKey)

	return nil
}
