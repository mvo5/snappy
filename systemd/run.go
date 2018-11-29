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

package systemd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/snapcore/snapd/strutil"
)

var strutilMakeRandomString = strutil.MakeRandomString

// RunWithOutput runs the given command "name" with the given args and returns
// the output and exit status.
func RunWithOutput(name string, arg ...string) ([]byte, error) {
	if os.Getuid() != 0 {
		return nil, fmt.Errorf("cannot use systemd-run as non-root")
	}

	unit := strutilMakeRandomString(20)

	// run the command in systemd-run
	args := []string{"--wait", fmt.Sprintf("--unit=%s.unit", unit), name}
	args = append(args, arg...)
	err := exec.Command("systemd-run", args...).Run()
	// then collect the output from journald
	output, err1 := exec.Command("journalctl", "-u", unit).CombinedOutput()
	if err1 != nil {
		err = fmt.Errorf("%s: cannot get output %s", err, err1)
	}
	return output, err
}
