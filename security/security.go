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
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/logger"
	"github.com/ubuntu-core/snappy/policy"
	"github.com/ubuntu-core/snappy/release"
	"github.com/ubuntu-core/snappy/snap"
)

var (
	// Note: these are true for ubuntu-core but perhaps not other flavors
	defaultPolicyGroups = []string{"network-client"}

	defaultTemplateName = "default"
)

// Profile returns the security profile string in the form of
// "snap_app-path_version" or an error
func Profile(m *snap.Info, appName, baseDir string) (string, error) {
	cleanedName := strings.Replace(appName, "/", "-", -1)
	if m.Type == snap.TypeFramework || m.Type == snap.TypeGadget {
		return fmt.Sprintf("%s_%s_%s", m.Name, cleanedName, m.Version), nil
	}

	origin, err := originFromBasedir(baseDir)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s_%s_%s", m.Name, origin, cleanedName, m.Version), nil
}

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

// PolicyDefinition is used to provide hand-crafted policy
type PolicyDefinition struct {
	AppArmor string `yaml:"apparmor" json:"apparmor"`
	Seccomp  string `yaml:"seccomp" json:"seccomp"`
}

type PolicyResult struct {
	ID *AppID

	AaPolicy string
	AaFn     string

	ScPolicy string
	ScFn     string
}

type AppID struct {
	AppID   string
	Pkgname string
	Appname string
	Version string
}

func NewAppID(appID string) (*AppID, error) {
	tmp := strings.Split(appID, "_")
	if len(tmp) != 3 {
		return nil, ErrInvalidAppID
	}
	id := AppID{
		AppID:   appID,
		Pkgname: tmp[0],
		Appname: tmp[1],
		Version: tmp[2],
	}
	return &id, nil
}

// TODO: once verified, reorganize all these
func (sa *AppID) AppArmorVars() string {
	aavars := fmt.Sprintf(`
# Specified profile variables
@{APP_APPNAME}="%s"
@{APP_ID_DBUS}="%s"
@{APP_PKGNAME_DBUS}="%s"
@{APP_PKGNAME}="%s"
@{APP_VERSION}="%s"
@{INSTALL_DIR}="{/snaps,/gadget}"
# Deprecated:
@{CLICK_DIR}="{/snaps,/gadget}"`, sa.Appname, dbusPath(sa.AppID), dbusPath(sa.Pkgname), sa.Pkgname, sa.Version)
	return aavars
}

const allowed = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

// Generate a string suitable for use in a DBus object
func dbusPath(s string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))

	for _, c := range []byte(s) {
		if strings.IndexByte(allowed, c) >= 0 {
			fmt.Fprintf(buf, "%c", c)
		} else {
			fmt.Fprintf(buf, "_%02x", c)
		}
	}

	return buf.String()
}

// securityPolicyType is a kind of securityPolicy, we currently
// have "apparmor" and "seccomp"
type securityPolicyType struct {
	name          string
	basePolicyDir string
}

var PolicyTypeAppArmor = securityPolicyType{
	name:          "apparmor",
	basePolicyDir: "/usr/share/apparmor/easyprof",
}

var PolicyTypeSeccomp = securityPolicyType{
	name:          "seccomp",
	basePolicyDir: "/usr/share/seccomp",
}

func (sp *securityPolicyType) PolicyDir() string {
	return filepath.Join(dirs.GlobalRootDir, sp.basePolicyDir)
}

func (sp *securityPolicyType) frameworkPolicyDir() string {
	frameworkPolicyDir := filepath.Join(policy.SecBase, sp.name)
	return filepath.Join(dirs.GlobalRootDir, frameworkPolicyDir)
}

// FindTemplate returns the security template content from the template name.
func (sp *securityPolicyType) FindTemplate(templateName string) (string, error) {
	if templateName == "" {
		templateName = defaultTemplateName
	}

	subdir := filepath.Join("templates", defaultPolicyVendor(), defaultPolicyVersion())
	systemTemplateDir := filepath.Join(sp.PolicyDir(), subdir, templateName)
	fwTemplateDir := filepath.Join(sp.frameworkPolicyDir(), "templates", templateName)

	// Read system and framwork policy, but always prefer system policy
	fns := []string{systemTemplateDir, fwTemplateDir}
	for _, fn := range fns {
		content, err := ioutil.ReadFile(fn)
		// it is ok if the file does not exists
		if os.IsNotExist(err) {
			continue
		}
		// but any other error is a failure
		if err != nil {
			return "", err
		}

		return string(content), nil
	}

	return "", &ErrPolicyNotFound{"template", sp, templateName}
}

// findSingleCap returns the security template content for a single
// security-cap.
func (sp *securityPolicyType) findSingleCap(capName, systemPolicyDir, fwPolicyDir string) ([]string, error) {
	found := false
	p := []string{}

	policyDirs := []string{systemPolicyDir, fwPolicyDir}
	for _, dir := range policyDirs {
		fn := filepath.Join(dir, capName)
		newCaps, err := readSingleCapFile(fn)
		// its ok if the file does not exist
		if os.IsNotExist(err) {
			continue
		}
		// but any other error is not ok
		if err != nil {
			return nil, err
		}
		p = append(p, newCaps...)
		found = true
		break
	}

	if found == false {
		return nil, &ErrPolicyNotFound{"cap", sp, capName}
	}

	return p, nil
}

// FindCaps returns the security template content for the given list
// of security-caps.
func (sp *securityPolicyType) FindCaps(caps []string, templateName string) ([]string, error) {
	// XXX: this is snappy specific, on other systems like the phone we may
	// want different defaults.
	if templateName == "" && caps == nil {
		caps = defaultPolicyGroups
	}

	// Nothing to find if caps is empty
	if len(caps) == 0 {
		return nil, nil
	}

	subdir := filepath.Join("policygroups", defaultPolicyVendor(), defaultPolicyVersion())
	parentDir := filepath.Join(sp.PolicyDir(), subdir)
	fwParentDir := filepath.Join(sp.frameworkPolicyDir(), "policygroups")

	var p []string
	for _, c := range caps {
		newCap, err := sp.findSingleCap(c, parentDir, fwParentDir)
		if err != nil {
			return nil, err
		}
		p = append(p, newCap...)
	}

	return p, nil
}

func defaultPolicyVendor() string {
	// FIXME: slightly ugly that we have to give a prefix here
	return fmt.Sprintf("ubuntu-%s", release.Get().Flavor)
}

func defaultPolicyVersion() string {
	// note that we can not use release.Get().Series here
	// because that will return "rolling" for the development
	// version but apparmor stores its templates under the
	// version number (e.g. 16.04) instead
	ver, err := release.ReadLsb()
	if err != nil {
		// when this happens we are in trouble
		panic(err)
	}
	return ver.Release
}

// helper for findSingleCap that implements readlines().
func readSingleCapFile(fn string) ([]string, error) {
	p := []string{}

	r, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	s := bufio.NewScanner(r)
	for s.Scan() {
		p = append(p, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}

	return p, nil
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
