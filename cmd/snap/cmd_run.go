// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
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
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/snap"
)

type cmdRun struct {
	Positional struct {
		SnapCmd string `positional-arg-name:"snapcmd" description:"the snap command run, e.g. hello-world.echo"`
	} `positional-args:"yes" required:"yes"`
	Command string `long:"command" description:"run a non-default command (like post-stop"`
}

func init() {
	addCommand("run",
		i18n.G("Run the given snap command"),
		i18n.G("Run the given snap command with the right environment"),
		func() flags.Commander {
			return &cmdRun{}
		})
}

func splitSnapCmd(snapCmd string) (snap, app string) {
	l := strings.SplitN(snapCmd, ".", 2)
	if len(l) < 2 {
		return l[0], ""
	}
	return l[0], l[1]
}

func (x *cmdRun) Execute(args []string) error {
	snapName, appName := splitSnapCmd(x.Positional.SnapCmd)

	// we need to get the revision here because once we are inside
	// the confinement its not available anymore
	snaps, err := Client().ListSnaps([]string{snapName})
	if err != nil {
		return err
	}
	if len(snaps) == 0 {
		return fmt.Errorf("cannot find snap %q", x.Positional.SnapCmd)
	}
	if len(snaps) > 1 {
		return fmt.Errorf("multiple snaps for %q: %d", x.Positional.SnapCmd, len(snaps))
	}
	sn := snaps[0]
	info, err := snap.ReadInfo(snapName, &snap.SideInfo{
		Revision: snap.R(sn.Revision.N),
	})
	if err != nil {
		return err
	}
	var app *snap.AppInfo
	for _, ai := range info.Apps {
		if ai.Name == appName {
			app = ai
			break
		}
	}
	if app == nil {
		return fmt.Errorf("cannot find app %q in %q", appName, snapName)
	}

	// FIXME: setup SNAP_USER_DATA_DIR here before running the confinement
	cmd := []string{
		"/usr/bin/ubuntu-core-launcher",
		app.SecurityTag(),
		app.SecurityTag(),
		"/usr/lib/snapd/snap-launch",
		snapName,
		appName,
		// FIXME: command needs to take "--command=post-stop" into
		//        account
		app.Command,
		sn.Revision.String(),
	}
	cmd = append(cmd, args...)
	return syscall.Exec(cmd[0], cmd, os.Environ())
}
