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
	"syscall"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/snap"
	"github.com/snapcore/snapd/snap/snapenv"
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

func snapLaunch() error {
	// FIXME: use proper parser
	snapName := os.Args[1]
	appName := os.Args[2]
	command := os.Args[3]
	revision := os.Args[4]
	args := os.Args[5:]

	info, err := snap.ReadInfo(snapName, &snap.SideInfo{
		Revision: snap.R(revision),
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
	cmd := filepath.Join(app.Snap.MountDir(), command)
	return syscall.Exec(cmd, args, env)
}
