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
	"path/filepath"
	"syscall"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snapenv"
)

type cmdRun struct {
	Positional struct {
		Snap    string `positional-arg-name:"snap" description:"the snap"`
		App     string `positional-arg-name:"app" description:"the app"`
		Command string `positional-arg-name:"cmd" description:"the command to run"`
	} `positional-args:"yes" required:"yes"`
	AfterSuidSetup bool   `description:"after-suid-setup"`
	Revision       string `description:"revision of the snap"`
}

func init() {
	addCommand("run",
		i18n.G("Run the given command with the right environment"),
		i18n.G("Run the given command with the right environment"),
		func() flags.Commander {
			return &cmdRun{}
		})
}

func copyEnv(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}

	return out
}

func (x *cmdRun) Execute(args []string) error {
	// "outer" part of `snap run` which invokes the launcher
	if !x.AfterSuidSetup {
		// we need to get the revision here because once we are inside
		// the confinement its not available anymore
		snaps, err := Client().ListSnaps([]string{x.Positional.Snap})
		if err != nil {
			return err
		}
		if len(snaps) == 0 {
			return fmt.Errorf("cannot find snap %q", x.Positional.Snap)
		}
		if len(snaps) > 1 {
			return fmt.Errorf("multiple snaps for %q: %d", x.Positional.Snap, len(snaps))
		}
		sn := snaps[0]

		// FIXME copied code
		info, err := snap.ReadInfo(x.Positional.Snap, &snap.SideInfo{
			Revision: snap.R(x.Revision),
		})
		if err != nil {
			return err
		}
		var app *snap.AppInfo
		for _, ai := range info.Apps {
			if ai.Name == x.Positional.App {
				app = ai
				break
			}
		}
		if app == nil {
			return fmt.Errorf("cannot find app %q in %q", x.Positional.App, x.Positional.Snap)
		}

		// FIXME: setup SNAP_USER_DATA_DIR here
		cmd := []string{
			"/usr/bin/ubuntu-core-launcher",
			app.SecurityTag(),
			app.SecurityTag(),
			"/usr/bin/snap", "run", "--after-suid-setup",
			"--revision", sn.Revision.String(),
			x.Positional.Snap,
			x.Positional.App,
			x.Positional.Command,
		}
		cmd = append(cmd, args...)
		return syscall.Exec(cmd[0], cmd, os.Environ())
	}

	// we are running inside the apparmor confinement at this point

	// we need some stuff we have not available from the snapd
	info, err := snap.ReadInfo(x.Positional.Snap, &snap.SideInfo{
		Revision: snap.R(x.Revision),
	})
	if err != nil {
		return err
	}
	var app *snap.AppInfo
	for _, ai := range info.Apps {
		if ai.Name == x.Positional.App {
			app = ai
			break
		}
	}
	if app == nil {
		return fmt.Errorf("cannot find app %q in %q", x.Positional.App, x.Positional.Snap)
	}

	// build wrapper env
	env := os.Environ()
	wrapperData := struct {
		App     *snap.AppInfo
		EnvVars string
		// XXX: needed by snapenv
		SnapName string
		SnapArch string
		SnapPath string
		Version  string
		Revision snap.Revision
		Home     string
	}{
		App: app,
		// XXX: needed by snapenv
		SnapName: app.Snap.Name(),
		SnapArch: arch.UbuntuArchitecture(),
		SnapPath: app.Snap.MountDir(),
		Version:  app.Snap.Version,
		Revision: app.Snap.Revision,
		Home:     "$HOME",
	}
	for _, envVar := range append(
		snapenv.GetBasicSnapEnvVars(wrapperData),
		snapenv.GetUserSnapEnvVars(wrapperData)...) {
		env = append(env, envVar)
	}

	// build the evnironment from the yaml
	appEnv := copyEnv(app.Snap.Environment)
	for k, v := range app.Environment {
		appEnv[k] = v
	}
	for k, v := range appEnv {
		env = append(env, fmt.Sprintf("%s=%s\n", k, v))
	}

	// run the command
	cmd := filepath.Join(app.Snap.MountDir(), x.Positional.Command)
	return syscall.Exec(cmd, args, env)
}
