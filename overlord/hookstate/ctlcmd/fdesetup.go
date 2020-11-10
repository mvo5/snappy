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

	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/i18n"
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
	// XXX: make "op" a type: "initial-setup", "update"
	Op string `json:"op"`

	Key     []byte `json:"key,omitempty"`
	KeyName string `json:"key-name,omitempty"`

	// XXX: not set yet
	// XXX2: do we need this to be a list? i.e. multiple models?
	// Model related fields, see secboot:SnapModel interface
	Series    string             `json:"series,omitempty"`
	BrandID   string             `json:"brand-id,omitempty"`
	Model     string             `json:"model,omitempty"`
	Grade     asserts.ModelGrade `json:"grade,omitempty"`
	SignKeyID string             `json:"sign-key-id,omitempty"`

	// XXX: LoadChains, KernelCmdline
}

func (c *fdeSetupRequestCommand) Execute(args []string) error {
	context := c.context()
	if context == nil {
		return fmt.Errorf("cannot  without a context")
	}
	context.Lock()
	defer context.Unlock()

	var js fdeSetupJSON
	if err := context.Get("fde-op", &js.Op); err != nil {
		return fmt.Errorf("cannot get fde op from context: %v", err)
	}
	if err := context.Get("fde-key", &js.Key); err != nil {
		return fmt.Errorf("cannot get fde key from context: %v", err)
	}
	if err := context.Get("fde-key-name", &js.KeyName); err != nil {
		return fmt.Errorf("cannot get fde key name from context: %v", err)
	}
	var model asserts.Model
	if err := context.Get("model", &model); err != nil {
		return fmt.Errorf("cannot get model from context: %v", err)
	}
	// XXX: make this a helper
	js.Series = model.Series()
	js.BrandID = model.BrandID()
	js.Model = model.Model()
	js.Grade = model.Grade()
	js.SignKeyID = model.SignKeyID()

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
The fde-setup-result command reads the result data from a FDE setup
from stdin.

    $ echo "sealed-key" | snapctl fde-setup-result
`)

func init() {
	addCommand("fde-setup-result", shortFdeSetupResultHelp, longFdeSetupResultHelp, func() command { return &fdeSetupResultCommand{} })
}

func (c *fdeSetupResultCommand) Execute(args []string) error {
	context := c.context()
	if context == nil {
		return fmt.Errorf("cannot  without a context")
	}
	context.Lock()
	defer context.Unlock()

	var sealedKey []byte
	if err := context.Get("stdin", &sealedKey); err != nil {
		return fmt.Errorf("cannot get key from stdin: %v", err)
	}
	if sealedKey == nil {
		return fmt.Errorf("no sealed key data found on stdin")
	}
	task, ok := context.Task()
	if !ok {
		return fmt.Errorf("internal error: fdeSetupResultCommand called without task")
	}
	task.Set("fde-sealed-key", sealedKey)

	return nil
}
