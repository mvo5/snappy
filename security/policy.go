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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/policy"
	"github.com/ubuntu-core/snappy/release"
)

var (
	// Note: these are true for ubuntu-core but perhaps not other flavors
	defaultPolicyGroups = []string{"network-client"}

	defaultTemplateName = "default"
)

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
