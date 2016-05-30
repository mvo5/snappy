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

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/snapcore/snapd/snap"
)

func main() {
	if err := snapLaunch(); err != nil {
		fmt.Printf("cannot snapLaunch: %s\n", err)
		os.Exit(1)
	}
}

func copyEnv(in map[string]string) map[string]string {
	out := make(map[string]string)
	for k, v := range in {
		out[k] = v
	}

	return out
}

func splitSnapCmd(snapCmd string) (snap, app string) {
	l := strings.SplitN(snapCmd, ".", 2)
	if len(l) < 2 {
		return l[0], ""
	}
	return l[0], l[1]
}

func snapLaunch() error {
	// FIXME: use proper parser
	snapCmd := os.Args[1]
	args := os.Args[2:]

	snapName, appName := splitSnapCmd(snapCmd)

	info, err := snap.ReadInfo(snapName, &snap.SideInfo{
		Revision: snap.R(os.Getenv("SNAP_REVISION")),
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

	// build the evnironment from the yamle
	env := os.Environ()
	appEnv := copyEnv(app.Snap.Environment)
	for k, v := range app.Environment {
		appEnv[k] = v
	}
	for k, v := range appEnv {
		env = append(env, fmt.Sprintf("%s=%s\n", k, v))
	}

	// run the command
	cmd := filepath.Join(app.Snap.MountDir(), app.Command)
	return syscall.Exec(cmd, args, env)
}
