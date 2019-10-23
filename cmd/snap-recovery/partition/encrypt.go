// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
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
package partition

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/snapcore/snapd/osutil"
)

// XXX: what is the right size here?
var masterKeySize = 64

func createKey(size int) ([]byte, error) {
	buffer := make([]byte, size)
	_, err := rand.Read(buffer)
	// On return, n == len(b) if and only if err == nil
	return buffer, err
}

// MakeEncrypted will create a encrypted device partition from the
// exiting partition. Note that it will manipulate the given DeviceStructure
// so that the node points to the right device
func MakeEncrypted(part *DeviceStructure, partLabel string) error {
	tempKeyFile, err := ioutil.TempFile("", "enc")
	if err != nil {
		return err
	}
	// use wipe() here
	defer os.Remove(tempKeyFile.Name())

	key, err := createKey(masterKeySize)
	if err != nil {
		return fmt.Errorf("cannot create key: %v", err)
	}

	// XXX: Ideally we shouldn't write this key, but cryptsetup
	// only reads the master key from a file.
	if _, err := tempKeyFile.Write(key); err != nil {
		return fmt.Errorf("cannot create key file: %s", err)
	}
	cmd := exec.Command("cryptsetup", "-q", "luksFormat", "--type", "luks2", "--pbkdf-memory", "1000", "--master-key-file", tempKeyFile.Name(), part.Node)
	cmd.Stdin = bytes.NewReader([]byte("\n"))
	if output, err := cmd.CombinedOutput(); err != nil {
		return osutil.OutputErr(output, fmt.Errorf("cannot format encrypted device: %s", err))
	}
	if output, err := exec.Command("cryptsetup", "open", "--master-key-file", tempKeyFile.Name(), part.Node, partLabel).CombinedOutput(); err != nil {
		return osutil.OutputErr(output, fmt.Errorf("cannot open encrypted device on %s: %s", part.Node, err))
	}

	// XXX: fugly, modify node so that it points to the right place for
	// the filesystem creation
	part.Node = fmt.Sprintf("/dev/mapper/%s", partLabel)

	return nil
}
