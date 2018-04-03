// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package configcore

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/systemd"
)

type sysdLogger struct{}

func (l *sysdLogger) Notify(status string) {
	fmt.Fprintf(Stderr, "sysd: %s\n", status)
}

// createOrRemoveServiceDisableFile creates or removes a stamp file to indicate
// if a service should be disabled
func createOrRemoveServiceDisableFile(serviceName, path, msg, value string) error {
	switch value {
	case "true":
		return ioutil.WriteFile(path, []byte(msg+"\n"), 0644)
	case "false":
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	default:
		return fmt.Errorf("option %q has invalid value %q", serviceName, value)
	}
}

// switchDisableSSHService handles the special case of disabling/enabling ssh
// service on core devices.
func switchDisableSSHService(name, value string) error {
	sysd := systemd.New(dirs.GlobalRootDir, &sysdLogger{})
	sshCanary := filepath.Join(dirs.GlobalRootDir, "/etc/ssh/sshd_not_to_be_run")
	if err := createOrRemoveServiceDisableFile(name, sshCanary, "SSH has been disabled by snapd system configuration", value); err != nil {
		return err
	}

	serviceName := name + ".service"
	switch value {
	case "true":
		return sysd.Stop(serviceName, 5*time.Minute)
	case "false":
		// Unmask both sshd.service and ssh.service and ignore the
		// errors, if any. This undoes the damage done by earlier
		// versions of snapd.
		sysd.Unmask("sshd.service")
		sysd.Unmask("ssh.service")
		return sysd.Start(serviceName)
	default:
		return fmt.Errorf("option %q has invalid value %q", serviceName, value)
	}
}

func switchGenericSystemdService(serviceName, value string) error {
	sysd := systemd.New(dirs.GlobalRootDir, &sysdLogger{})

	switch value {
	case "true":
		if err := sysd.Disable(serviceName); err != nil {
			return err
		}
		if err := sysd.Mask(serviceName); err != nil {
			return err
		}
		return sysd.Stop(serviceName, 5*time.Minute)
	case "false":
		if err := sysd.Unmask(serviceName); err != nil {
			return err
		}
		if err := sysd.Enable(serviceName); err != nil {
			return err
		}
		return sysd.Start(serviceName)
	default:
		return fmt.Errorf("option %q has invalid value %q", serviceName, value)
	}
}

// switchDisableService switches a service in/out of disabled state
// where "true" means disabled and "false" means enabled.
func switchDisableService(name, value string) error {
	switch name {
	case "ssh":
		return switchDisableSSHService(name, value)
	case "console-conf":
		disableStampPath := filepath.Join(dirs.GlobalRootDir, "/var/lib/console-conf/complete")
		return createOrRemoveServiceDisableFile(name, disableStampPath, "console-conf has been disabled by snapd system configuration", value)
	case "rsyslog":
		return switchGenericSystemdService(name+".service", value)
	default:
		return fmt.Errorf("trying to disable unsupported service %q", name)
	}
}

// services that can be disabled
func handleServiceDisableConfiguration(tr Conf) error {
	for _, service := range []string{"ssh", "rsyslog", "console-conf"} {
		output, err := coreCfg(tr, fmt.Sprintf("service.%s.disable", service))
		if err != nil {
			return err
		}
		if output != "" {
			if err := switchDisableService(service, output); err != nil {
				return err
			}
		}
	}

	return nil
}
