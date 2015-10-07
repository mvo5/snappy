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

package autopkgtest

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"launchpad.net/snappy/_integration-tests/testutils/tlog"
)

const (
	commonSSHOptions = "--- ssh "
	sshTimeout       = 600
)

var (
	// dependency aliasing
	tlogGetLevel = tlog.GetLevel
)

func kvmSSHOptions(imagePath string) string {
	var showBoot string

	if tlogGetLevel() == tlog.DebugLevel {
		showBoot = " -b"
	}
	return fmt.Sprintf(commonSSHOptions+
		"-s /usr/share/autopkgtest/ssh-setup/snappy --%s -i "+imagePath, showBoot)
}

func remoteTestbedSSHOptions(testbedIP string, testbedPort int) string {
	return fmt.Sprint(commonSSHOptions,
		"-H ", testbedIP,
		" -p ", strconv.Itoa(testbedPort),
		" -l ubuntu",
		" -i ", filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"),
		" --reboot",
		" --timeout-ssh ", strconv.Itoa(sshTimeout))
}
