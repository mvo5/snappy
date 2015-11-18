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

package snappy

import (
	"fmt"
	"os/exec"
)

// isMounted checks if the given dir is a mountpoint
func isMounted(dir string) bool {
	err := exec.Command("mountpoint", dir).Run()
	return err == nil
}

// lazyUnmount will unmount the given mountpoint
func lazyUnmount(mp string) error {
	output, err := exec.Command("umount", "--lazy", mp).CombinedOutput()
	if err != nil {
		return fmt.Errorf("umount failed with %s: %q", err, output)
	}

	return nil
}
