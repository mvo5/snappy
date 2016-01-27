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

package security

import (
	"github.com/ubuntu-core/snappy/logger"
)

// OverrideDefinition is used to override apparmor or seccomp
// security defaults
type OverrideDefinition struct {
	ReadPaths    []string `yaml:"read-paths,omitempty" json:"read-paths,omitempty"`
	WritePaths   []string `yaml:"write-paths,omitempty" json:"write-paths,omitempty"`
	Abstractions []string `yaml:"abstractions,omitempty" json:"abstractions,omitempty"`
	Syscalls     []string `yaml:"syscalls,omitempty" json:"syscalls,omitempty"`

	// deprecated keys, we warn when we see those
	DeprecatedAppArmor interface{} `yaml:"apparmor,omitempty" json:"apparmor,omitempty"`
	DeprecatedSeccomp  interface{} `yaml:"seccomp,omitempty" json:"seccomp,omitempty"`
}

// Definitions contains the common apparmor/seccomp definitions
type Definitions struct {
	// SecurityTemplate is a template name like "default"
	SecurityTemplate string `yaml:"security-template,omitempty" json:"security-template,omitempty"`
	// SecurityOverride is a override for the high level security json
	SecurityOverride *OverrideDefinition `yaml:"security-override,omitempty" json:"security-override,omitempty"`
	// SecurityPolicy is a hand-crafted low-level policy
	SecurityPolicy *PolicyDefinition `yaml:"security-policy,omitempty" json:"security-policy,omitempty"`

	// SecurityCaps is are the apparmor/seccomp capabilities for an app
	SecurityCaps []string `yaml:"caps,omitempty" json:"caps,omitempty"`
}

// NeedsAppArmorUpdate checks whether the security definitions are impacted by
// changes to policies or templates.
func (sd *Definitions) NeedsAppArmorUpdate(policies, templates map[string]bool) bool {
	if sd.SecurityPolicy != nil {
		return false
	}

	if sd.SecurityOverride != nil {
		// XXX: actually inspect the override to figure out in more detail
		return true
	}

	if templates[sd.SecurityTemplate] {
		return true
	}

	for _, cap := range sd.SecurityCaps {
		if policies[cap] {
			return true
		}
	}

	return false
}
func (sd *Definitions) MergeAppArmorSecurityOverrides(new *OverrideDefinition) {
	// nothing to do
	if new == nil {
		return
	}

	// ensure we have valid structs to work with
	if sd.SecurityOverride == nil {
		sd.SecurityOverride = &OverrideDefinition{}
	}

	sd.SecurityOverride.ReadPaths = append(sd.SecurityOverride.ReadPaths, new.ReadPaths...)
	sd.SecurityOverride.WritePaths = append(sd.SecurityOverride.WritePaths, new.WritePaths...)
	sd.SecurityOverride.Abstractions = append(sd.SecurityOverride.Abstractions, new.Abstractions...)
}

func (sd *Definitions) WarnDeprecatedKeys() {
	if sd.SecurityOverride != nil && sd.SecurityOverride.DeprecatedAppArmor != nil {
		logger.Noticef("The security-override.apparmor key is no longer supported, please use use security-override directly")
	}
	if sd.SecurityOverride != nil && sd.SecurityOverride.DeprecatedSeccomp != nil {
		logger.Noticef("The security-override.seccomp key is no longer supported, please use use security-override directly")
	}
}
