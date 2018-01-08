// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package main

import (
	"encoding/json"
	"fmt"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/advisor"
	"github.com/snapcore/snapd/i18n"
)

type cmdAdviseCommand struct {
	Positionals struct {
		Command string `required:"yes"`
	} `positional-args:"true"`

	Format string `long:"format" default:"pretty"`
}

var shortAdviseCommandHelp = i18n.G("Advise on available snaps.")
var longAdviseCommandHelp = i18n.G(`
The advise-command command shows what snaps with the given command are 
available.
`)

func init() {
	cmd := addCommand("advise-command", shortAdviseCommandHelp, longAdviseCommandHelp, func() flags.Commander {
		return &cmdAdviseCommand{}
	}, nil, []argDesc{
		{name: "<command>"},
	})
	cmd.hidden = true
}

func outputAdviseExactText(command string, result []advisor.Command) error {
	fmt.Fprintf(Stdout, i18n.G("The program %q can be found in the following snaps:\n"), command)
	for _, snap := range result {
		fmt.Fprintf(Stdout, " * %s\n", snap.Snap)
	}
	fmt.Fprintf(Stdout, i18n.G("Try: snap install <selected snap>\n"))
	return nil
}

func outputAdviseMisspellText(command string, result []advisor.Command) error {
	fmt.Fprintf(Stdout, i18n.G("No command %q found, did you mean:\n"), command)
	for _, snap := range result {
		fmt.Fprintf(Stdout, " Command %q from snap %q\n", snap.Command, snap.Snap)
	}
	return nil
}

func outputAdviseJSON(command string, results []advisor.Command) error {
	enc := json.NewEncoder(Stdout)
	enc.Encode(results)
	return nil
}

func (x *cmdAdviseCommand) Execute(args []string) error {
	if len(args) > 0 {
		return ErrExtraArgs
	}

	return adviseCommand(x.Positionals.Command, x.Format)
}

func adviseCommand(cmd string, format string) error {
	// find exact matches
	matches, err := advisor.FindCommand(cmd)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		switch format {
		case "json":
			return outputAdviseJSON(cmd, matches)
		case "pretty":
			return outputAdviseExactText(cmd, matches)
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	}

	// find misspellings
	matches, err = advisor.FindMisspelledCommand(cmd)
	if err != nil {
		return err
	}
	if len(matches) > 0 {
		switch format {
		case "json":
			return outputAdviseJSON(cmd, matches)
		case "pretty":
			return outputAdviseMisspellText(cmd, matches)
		default:
			return fmt.Errorf("unsupported format %q", format)
		}
	}

	return nil
}
