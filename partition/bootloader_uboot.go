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

package partition

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/helpers"

	"github.com/mvo5/goconfigparser"
	"github.com/mvo5/uboot-go/uenv"
)

const (
	bootloaderUbootDirReal        = "/boot/uboot"
	bootloaderUbootConfigFileReal = "uEnv.txt"

	// File created by u-boot itself when
	// bootloaderBootmodeTry == "try" which the
	// successfully booted system must remove to flag to u-boot that
	// this partition is "good".
	bootloaderUbootStampFileReal = "snappy-stamp.txt"

	// DEPRECATED:
	bootloaderUbootEnvFileReal = "snappy-system.txt"

	// the real uboot env
	bootloaderUbootFwEnvFileReal = "uboot.env"
)

// var to make it testable
var (
	atomicWriteFile = helpers.AtomicWriteFile
)

const bootloaderNameUboot bootloaderName = "u-boot"

type uboot struct {
	// set to true if the legacy uboot environemnt text file
	// needs to be used
	useLegacy bool
}

// Stores a Name and a Value to be added as a name=value pair in a file.
// TODO convert to map
type configFileChange struct {
	Name  string
	Value string
}

func bootloaderUbootDir() string {
	return filepath.Join(dirs.GlobalRootDir, bootloaderUbootDirReal)
}

func bootloaderUbootConfigFile() string {
	return filepath.Join(bootloaderUbootDir(), bootloaderUbootConfigFileReal)
}

func bootloaderUbootStampFile() string {
	return filepath.Join(bootloaderUbootDir(), bootloaderUbootStampFileReal)
}

func bootloaderUbootEnvFile() string {
	return filepath.Join(bootloaderUbootDir(), bootloaderUbootEnvFileReal)
}

func bootloaderUbootFwEnvFile() string {
	return filepath.Join(bootloaderUbootDir(), bootloaderUbootFwEnvFileReal)
}

// newUboot create a new Uboot bootloader object
func newUboot() bootLoader {
	if !helpers.FileExists(bootloaderUbootConfigFile()) {
		return nil
	}

	u := uboot{}
	if !helpers.FileExists(bootloaderUbootFwEnvFile()) {
		u.useLegacy = true
	}

	return &u
}

func (u *uboot) Name() bootloaderName {
	return bootloaderNameUboot
}

func getBootVarLegacy(name string) (value string, err error) {
	cfg := goconfigparser.New()
	cfg.AllowNoSectionHeader = true
	if err := cfg.ReadFile(bootloaderUbootEnvFile()); err != nil {
		return "", nil
	}

	return cfg.Get("", name)
}

func setBootVarLegacy(name, value string) error {
	curVal, err := getBootVarLegacy(name)
	if err == nil && curVal == value {
		return nil
	}

	changes := []configFileChange{
		configFileChange{
			Name:  name,
			Value: value,
		},
	}

	return modifyNameValueFile(bootloaderUbootEnvFile(), changes)
}

func setBootVarFwEnv(name, value string) error {
	env, err := uenv.Open(bootloaderUbootFwEnvFile())
	if err != nil {
		return err
	}

	// already set, nothing to do
	if env.Get(name) == value {
		return nil
	}

	env.Set(name, value)
	return env.Save()
}

func getBootVarFwEnv(name string) (string, error) {
	env, err := uenv.Open(bootloaderUbootFwEnvFile())
	if err != nil {
		return "", err
	}

	return env.Get(name), nil
}

func (u *uboot) GetBootVar(name string) (value string, err error) {
	if u.useLegacy {
		return getBootVarLegacy(name)
	}

	return getBootVarFwEnv(name)
}

func (u *uboot) SetBootVar(name, value string) error {
	if u.useLegacy {
		return setBootVarLegacy(name, value)
	}

	return setBootVarFwEnv(name, value)
}

func (u *uboot) BootDir() string {
	return bootloaderUbootDir()
}

// Rewrite the specified file, applying the specified set of changes.
// Lines not in the changes slice are left alone.
// If the original file does not contain any of the name entries (from
// the corresponding configFileChange objects), those entries are
// appended to the file.
//
// FIXME: put into utils package
// FIXME: improve logic
func modifyNameValueFile(path string, changes []configFileChange) error {
	var updated []configFileChange

	// we won't write to a file if we don't need to.
	updateNeeded := false

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, change := range changes {
			if strings.HasPrefix(line, fmt.Sprintf("%s=", change.Name)) {
				value := strings.SplitN(line, "=", 2)[1]
				// updated is used later to see if you had the originally requested
				// value.
				updated = append(updated, change)
				if value != change.Value {
					line = fmt.Sprintf("%s=%s", change.Name, change.Value)
					updateNeeded = true
				}
			}
		}
		if _, err := fmt.Fprintln(buf, line); err != nil {
			return err
		}
	}

	for _, change := range changes {
		got := false
		for _, update := range updated {
			if update.Name == change.Name {
				got = true
				break
			}
		}

		if !got {
			updateNeeded = true

			// name/value pair did not exist in original
			// file, so append
			if _, err := fmt.Fprintf(buf, "%s=%s\n", change.Name, change.Value); err != nil {
				return err
			}
		}
	}

	if updateNeeded {
		return atomicWriteFile(path, buf.Bytes(), 0644, 0)
	}

	return nil
}
