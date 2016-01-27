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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/helpers"
	"github.com/ubuntu-core/snappy/logger"
	"github.com/ubuntu-core/snappy/security"
	"github.com/ubuntu-core/snappy/snap"
	"github.com/ubuntu-core/snappy/snap/app"
)

var (

	// AppArmor cache dir
	aaCacheDir = "/var/cache/apparmor"

	errOriginNotFound     = errors.New("could not detect origin")
	errPolicyTypeNotFound = errors.New("could not find specified policy type")
	errPolicyGen          = errors.New("errors found when generating policy")

	// snappyConfig is the default securityDefinition for a snappy
	// config fragment
	snappyConfig = &SecurityDefinitions{
		SecurityCaps: []string{},
	}

	runAppArmorParser = runAppArmorParserImpl
)

func runAppArmorParserImpl(argv ...string) ([]byte, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	return cmd.CombinedOutput()
}

// SecurityDefinitions contains the common apparmor/seccomp definitions
type SecurityDefinitions struct {
	// SecurityTemplate is a template name like "default"
	SecurityTemplate string `yaml:"security-template,omitempty" json:"security-template,omitempty"`
	// SecurityOverride is a override for the high level security json
	SecurityOverride *security.OverrideDefinition `yaml:"security-override,omitempty" json:"security-override,omitempty"`
	// SecurityPolicy is a hand-crafted low-level policy
	SecurityPolicy *security.PolicyDefinition `yaml:"security-policy,omitempty" json:"security-policy,omitempty"`

	// SecurityCaps is are the apparmor/seccomp capabilities for an app
	SecurityCaps []string `yaml:"caps,omitempty" json:"caps,omitempty"`
}

// NeedsAppArmorUpdate checks whether the security definitions are impacted by
// changes to policies or templates.
func (sd *SecurityDefinitions) NeedsAppArmorUpdate(policies, templates map[string]bool) bool {
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

// Calculate whitespace prefix based on occurrence of s in t
func findWhitespacePrefix(t string, s string) string {
	subs := regexp.MustCompile(`(?m)^( *)` + regexp.QuoteMeta(s)).FindStringSubmatch(t)
	if subs == nil {
		return ""
	}

	return subs[1]
}

func genAppArmorPathRule(path string, access string) (string, error) {
	if !strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "@{") {
		logger.Noticef("Bad path: %s", path)
		return "", errPolicyGen
	}

	owner := ""
	if strings.HasPrefix(path, "/home") || strings.HasPrefix(path, "@{HOME") {
		owner = "owner "
	}

	rules := ""
	if strings.HasSuffix(path, "/") {
		rules += fmt.Sprintf("%s %s,\n", path, access)
		rules += fmt.Sprintf("%s%s** %s,\n", owner, path, access)
	} else if strings.HasSuffix(path, "/**") || strings.HasSuffix(path, "/*") {
		rules += fmt.Sprintf("%s/ %s,\n", filepath.Dir(path), access)
		rules += fmt.Sprintf("%s%s %s,\n", owner, path, access)
	} else {
		rules += fmt.Sprintf("%s%s %s,\n", owner, path, access)
	}

	return rules, nil
}

func mergeAppArmorTemplateAdditionalContent(appArmorTemplate, aaPolicy string, overrides *security.OverrideDefinition) (string, error) {
	// ensure we have
	if overrides == nil {
		overrides = &security.OverrideDefinition{}
	}

	if overrides.ReadPaths == nil {
		aaPolicy = strings.Replace(aaPolicy, "###READS###\n", "# No read paths specified\n", 1)
	} else {
		s := "# Additional read-paths from security-override\n"
		prefix := findWhitespacePrefix(appArmorTemplate, "###READS###")
		for _, readpath := range overrides.ReadPaths {
			rules, err := genAppArmorPathRule(strings.Trim(readpath, " "), "rk")
			if err != nil {
				return "", err
			}
			lines := strings.Split(rules, "\n")
			for _, rule := range lines {
				s += fmt.Sprintf("%s%s\n", prefix, rule)
			}
		}
		aaPolicy = strings.Replace(aaPolicy, "###READS###\n", s, 1)
	}

	if overrides.WritePaths == nil {
		aaPolicy = strings.Replace(aaPolicy, "###WRITES###\n", "# No write paths specified\n", 1)
	} else {
		s := "# Additional write-paths from security-override\n"
		prefix := findWhitespacePrefix(appArmorTemplate, "###WRITES###")
		for _, writepath := range overrides.WritePaths {
			rules, err := genAppArmorPathRule(strings.Trim(writepath, " "), "rwk")
			if err != nil {
				return "", err
			}
			lines := strings.Split(rules, "\n")
			for _, rule := range lines {
				s += fmt.Sprintf("%s%s\n", prefix, rule)
			}
		}
		aaPolicy = strings.Replace(aaPolicy, "###WRITES###\n", s, 1)
	}

	if overrides.Abstractions == nil {
		aaPolicy = strings.Replace(aaPolicy, "###ABSTRACTIONS###\n", "# No abstractions specified\n", 1)
	} else {
		s := "# Additional abstractions from security-override\n"
		prefix := findWhitespacePrefix(appArmorTemplate, "###ABSTRACTIONS###")
		for _, abs := range overrides.Abstractions {
			s += fmt.Sprintf("%s#include <abstractions/%s>\n", prefix, abs)
		}
		aaPolicy = strings.Replace(aaPolicy, "###ABSTRACTIONS###\n", s, 1)
	}

	return aaPolicy, nil
}

func getAppArmorTemplatedPolicy(m *snap.Info, appID *security.AppID, template string, caps []string, overrides *security.OverrideDefinition) (string, error) {
	t, err := security.PolicyTypeAppArmor.FindTemplate(template)
	if err != nil {
		return "", err
	}
	p, err := security.PolicyTypeAppArmor.FindCaps(caps, template)
	if err != nil {
		return "", err
	}

	aaPolicy := strings.Replace(t, "\n###VAR###\n", appID.AppArmorVars()+"\n", 1)
	aaPolicy = strings.Replace(aaPolicy, "\n###PROFILEATTACH###", fmt.Sprintf("\nprofile \"%s\"", appID.AppID), 1)

	aacaps := ""
	if len(p) == 0 {
		aacaps += "# No caps (policy groups) specified\n"
	} else {
		aacaps += "# Rules specified via caps (policy groups)\n"
		prefix := findWhitespacePrefix(t, "###POLICYGROUPS###")
		for _, line := range p {
			if len(line) == 0 {
				aacaps += "\n"
			} else {
				aacaps += fmt.Sprintf("%s%s\n", prefix, line)
			}
		}
	}
	aaPolicy = strings.Replace(aaPolicy, "###POLICYGROUPS###\n", aacaps, 1)

	return mergeAppArmorTemplateAdditionalContent(t, aaPolicy, overrides)
}

func getSeccompTemplatedPolicy(m *snap.Info, appID *security.AppID, templateName string, caps []string, overrides *security.OverrideDefinition) (string, error) {
	t, err := security.PolicyTypeSeccomp.FindTemplate(templateName)
	if err != nil {
		return "", err
	}
	p, err := security.PolicyTypeSeccomp.FindCaps(caps, templateName)
	if err != nil {
		return "", err
	}

	scPolicy := t + "\n" + strings.Join(p, "\n")

	if overrides != nil && overrides.Syscalls != nil {
		scPolicy += "\n# Addtional syscalls from security-override\n"
		for _, syscall := range overrides.Syscalls {
			scPolicy += fmt.Sprintf("%s\n", syscall)
		}
	}

	scPolicy = strings.Replace(scPolicy, "\ndeny ", "\n# EXPLICITLY DENIED: ", -1)

	return scPolicy, nil
}

var finalCurtain = regexp.MustCompile(`}\s*$`)

func getAppArmorCustomPolicy(m *snap.Info, appID *security.AppID, fn string, overrides *security.OverrideDefinition) (string, error) {
	custom, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	aaPolicy := strings.Replace(string(custom), "\n###VAR###\n", appID.AppArmorVars()+"\n", 1)
	aaPolicy = strings.Replace(aaPolicy, "\n###PROFILEATTACH###", fmt.Sprintf("\nprofile \"%s\"", appID.AppID), 1)

	// a custom policy may not have the overrides defined that we
	// use for the hw-assign work. so we insert them here
	aaPolicy = finalCurtain.ReplaceAllString(aaPolicy, `
###READS###
###WRITES###
###ABSTRACTIONS###
}
`)

	return mergeAppArmorTemplateAdditionalContent("", aaPolicy, overrides)
}

func getSeccompCustomPolicy(m *snap.Info, appID *security.AppID, fn string) (string, error) {
	custom, err := ioutil.ReadFile(fn)
	if err != nil {
		return "", err
	}

	return string(custom), nil
}

var loadAppArmorPolicy = func(fn string) ([]byte, error) {
	args := []string{
		"/sbin/apparmor_parser",
		"-r",
		"--write-cache",
		"-L", aaCacheDir,
		fn,
	}
	content, err := runAppArmorParser(args...)
	if err != nil {
		logger.Noticef("%v failed", args)
	}
	return content, err
}

func removeOneSecurityPolicy(m *snap.Info, name, baseDir string) error {
	profileName, err := security.Profile(m, filepath.Base(name), baseDir)
	if err != nil {
		return err
	}

	// seccomp profile
	fn := filepath.Join(dirs.SnapSeccompDir, profileName)
	if err := os.Remove(fn); err != nil && !os.IsNotExist(err) {
		return err
	}

	// apparmor cache
	fn = filepath.Join(aaCacheDir, profileName)
	if err := os.Remove(fn); err != nil && !os.IsNotExist(err) {
		return err
	}

	// apparmor profile
	fn = filepath.Join(dirs.SnapAppArmorDir, profileName)
	if err := os.Remove(fn); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func removePolicy(m *snap.Info, apps map[string]*app.Yaml, baseDir string) error {
	for _, app := range apps {
		if err := removeOneSecurityPolicy(m, app.Name, baseDir); err != nil {
			return err
		}
	}

	if err := removeOneSecurityPolicy(m, "snappy-config", baseDir); err != nil {
		return err
	}

	return nil
}

func (sd *SecurityDefinitions) mergeAppArmorSecurityOverrides(new *security.OverrideDefinition) {
	// nothing to do
	if new == nil {
		return
	}

	// ensure we have valid structs to work with
	if sd.SecurityOverride == nil {
		sd.SecurityOverride = &security.OverrideDefinition{}
	}

	sd.SecurityOverride.ReadPaths = append(sd.SecurityOverride.ReadPaths, new.ReadPaths...)
	sd.SecurityOverride.WritePaths = append(sd.SecurityOverride.WritePaths, new.WritePaths...)
	sd.SecurityOverride.Abstractions = append(sd.SecurityOverride.Abstractions, new.Abstractions...)
}

func (sd *SecurityDefinitions) warnDeprecatedKeys() {
	if sd.SecurityOverride != nil && sd.SecurityOverride.DeprecatedAppArmor != nil {
		logger.Noticef("The security-override.apparmor key is no longer supported, please use use security-override directly")
	}
	if sd.SecurityOverride != nil && sd.SecurityOverride.DeprecatedSeccomp != nil {
		logger.Noticef("The security-override.seccomp key is no longer supported, please use use security-override directly")
	}
}

func (sd *SecurityDefinitions) generatePolicyForServiceBinaryResult(m *snap.Info, name string, baseDir string) (*security.PolicyResult, error) {
	res := &security.PolicyResult{}
	appID, err := security.Profile(m, name, baseDir)
	if err != nil {
		logger.Noticef("Failed to obtain security profile for %s: %v", name, err)
		return nil, err
	}

	res.ID, err = security.NewAppID(appID)
	if err != nil {
		logger.Noticef("Failed to obtain APP_ID for %s: %v", name, err)
		return nil, err
	}

	// warn about deprecated
	sd.warnDeprecatedKeys()

	// add the hw-override parts and merge with the other overrides
	origin := ""
	if m.Type != snap.TypeFramework && m.Type != snap.TypeGadget {
		origin, err = originFromYamlPath(filepath.Join(baseDir, "meta", "snap.yaml"))
		if err != nil {
			return nil, err
		}
	}

	hwaccessOverrides, err := readHWAccessYamlFile(qualifiedName(m, origin))
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	sd.mergeAppArmorSecurityOverrides(&hwaccessOverrides)
	if sd.SecurityPolicy != nil {
		res.AaPolicy, err = getAppArmorCustomPolicy(m, res.ID, filepath.Join(baseDir, sd.SecurityPolicy.AppArmor), sd.SecurityOverride)
		if err != nil {
			logger.Noticef("Failed to generate custom AppArmor policy for %s: %v", name, err)
			return nil, err
		}
		res.ScPolicy, err = getSeccompCustomPolicy(m, res.ID, filepath.Join(baseDir, sd.SecurityPolicy.Seccomp))
		if err != nil {
			logger.Noticef("Failed to generate custom seccomp policy for %s: %v", name, err)
			return nil, err
		}
	} else {
		res.AaPolicy, err = getAppArmorTemplatedPolicy(m, res.ID, sd.SecurityTemplate, sd.SecurityCaps, sd.SecurityOverride)
		if err != nil {
			logger.Noticef("Failed to generate AppArmor policy for %s: %v", name, err)
			return nil, err
		}

		res.ScPolicy, err = getSeccompTemplatedPolicy(m, res.ID, sd.SecurityTemplate, sd.SecurityCaps, sd.SecurityOverride)
		if err != nil {
			logger.Noticef("Failed to generate seccomp policy for %s: %v", name, err)
			return nil, err
		}
	}
	res.ScFn = filepath.Join(dirs.SnapSeccompDir, res.ID.AppID)
	res.AaFn = filepath.Join(dirs.SnapAppArmorDir, res.ID.AppID)

	return res, nil
}

func (sd *SecurityDefinitions) generatePolicyForServiceBinary(m *snap.Info, name string, baseDir string) error {
	p, err := sd.generatePolicyForServiceBinaryResult(m, name, baseDir)
	if err != nil {
		return err
	}

	os.MkdirAll(filepath.Dir(p.ScFn), 0755)
	err = helpers.AtomicWriteFile(p.ScFn, []byte(p.ScPolicy), 0644, 0)
	if err != nil {
		logger.Noticef("Failed to write seccomp policy for %s: %v", name, err)
		return err
	}

	os.MkdirAll(filepath.Dir(p.AaFn), 0755)
	err = helpers.AtomicWriteFile(p.AaFn, []byte(p.AaPolicy), 0644, 0)
	if err != nil {
		logger.Noticef("Failed to write AppArmor policy for %s: %v", name, err)
		return err
	}
	out, err := loadAppArmorPolicy(p.AaFn)
	if err != nil {
		logger.Noticef("Failed to load AppArmor policy for %s: %v\n:%s", name, err, out)
		return err
	}

	return nil
}

// FIXME: move into something more generic - SnapPart.HasConfig?
func hasConfig(baseDir string) bool {
	return helpers.FileExists(filepath.Join(baseDir, "meta", "hooks", "config"))
}

func findSkillForApp(m *snapYaml, app *app.Yaml) (*usesYaml, error) {
	if len(app.UsesRef) == 0 {
		return nil, nil
	}
	if len(app.UsesRef) != 1 {
		return nil, fmt.Errorf("only a single skill is supported, %d found", len(app.UsesRef))
	}

	skill, ok := m.Uses[app.UsesRef[0]]
	if !ok {
		return nil, fmt.Errorf("can not find skill %q", app.UsesRef[0])
	}
	return skill, nil
}

func generatePolicy(m *snapYaml, baseDir string) error {
	var foundError error

	// generate default security config for snappy-config
	if hasConfig(baseDir) {
		if err := snappyConfig.generatePolicyForServiceBinary(m.info(), "snappy-config", baseDir); err != nil {
			foundError = err
			logger.Noticef("Failed to obtain APP_ID for %s: %v", "snappy-config", err)
		}
	}

	for _, app := range m.Apps {
		skill, err := findSkillForApp(m, app)
		if err != nil {
			return err
		}
		if skill == nil {
			continue
		}

		err = skill.generatePolicyForServiceBinary(m.info(), app.Name, baseDir)
		if err != nil {
			foundError = err
			logger.Noticef("Failed to generate policy for service %s: %v", app.Name, err)
			continue
		}
	}

	// FIXME: if there are multiple errors only the last one
	//        will be preserved
	if foundError != nil {
		return foundError
	}

	return nil
}

// regeneratePolicyForSnap is used to regenerate all security policy for a
// given snap
func regeneratePolicyForSnap(snapname string) error {
	globExpr := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("%s_*", snapname))
	matches, err := filepath.Glob(globExpr)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		// Nothing to regenerate is not an error
		return nil
	}

	appliedVersion := ""
	for _, profile := range matches {
		appID, err := security.NewAppID(filepath.Base(profile))
		if err != nil {
			return err
		}
		if appID.Version != appliedVersion {
			// FIXME: dirs.SnapSnapsDir is too simple, gadget
			fn := filepath.Join(dirs.SnapSnapsDir, appID.Pkgname, appID.Version, "meta", "snap.yaml")
			if !helpers.FileExists(fn) {
				continue
			}
			err := GeneratePolicyFromFile(fn, true)
			if err != nil {
				return err
			}
			appliedVersion = appID.Version
		}
	}

	return nil
}

// compare if the given policy matches the current system policy
// return an error if not
func comparePolicyToCurrent(p *security.PolicyResult) error {
	if err := compareSinglePolicyToCurrent(p.AaFn, p.AaPolicy); err != nil {
		return err
	}
	if err := compareSinglePolicyToCurrent(p.ScFn, p.ScPolicy); err != nil {
		return err
	}

	return nil
}

// helper for comparePolicyToCurrent that takes a single apparmor or seccomp
// policy and compares it to the system version
func compareSinglePolicyToCurrent(oldPolicyFn, newPolicy string) error {
	oldPolicy, err := ioutil.ReadFile(oldPolicyFn)
	if err != nil {
		return err
	}
	if string(oldPolicy) != newPolicy {
		return fmt.Errorf("policy differs %s", oldPolicyFn)
	}
	return nil
}

// CompareGeneratePolicyFromFile is used to simulate security policy
// generation and returns if the policy would have changed
func CompareGeneratePolicyFromFile(fn string) error {
	m, err := parseSnapYamlFileWithVersion(fn)
	if err != nil {
		return err
	}

	baseDir := filepath.Dir(filepath.Dir(fn))

	for _, app := range m.Apps {
		skill, err := findSkillForApp(m, app)
		if err != nil {
			return err
		}
		if skill == nil {
			continue
		}

		p, err := skill.generatePolicyForServiceBinaryResult(m.info(), app.Name, baseDir)
		// FIXME: use apparmor_profile -p on both AppArmor profiles
		if err != nil {
			// FIXME: what to do here?
			return err
		}
		if err := comparePolicyToCurrent(p); err != nil {
			return err
		}
	}

	// now compare the snappy-config profile
	if hasConfig(baseDir) {
		p, err := snappyConfig.generatePolicyForServiceBinaryResult(m.info(), "snappy-config", baseDir)
		if err != nil {
			return nil
		}
		if err := comparePolicyToCurrent(p); err != nil {
			return err
		}
	}

	return nil
}

// FIXME: refactor so that we don't need this
func parseSnapYamlFileWithVersion(fn string) (*snapYaml, error) {
	m, err := parseSnapYamlFile(fn)

	// FIXME: duplicated code from snapp.go:NewSnapPartFromYaml,
	//        version is overriden by sideloaded versions

	// use EvalSymlinks to resolve 'current' to the correct version
	dir, err := filepath.EvalSymlinks(filepath.Dir(filepath.Dir(fn)))
	if err != nil {
		return nil, err
	}
	m.Version = filepath.Base(dir)

	return m, err
}

// GeneratePolicyFromFile is used to generate security policy on the system
// from the specified manifest file name
func GeneratePolicyFromFile(fn string, force bool) error {
	// FIXME: force not used yet
	m, err := parseSnapYamlFileWithVersion(fn)
	if err != nil {
		return err
	}

	if m.Type == "" || m.Type == snap.TypeApp {
		_, err = originFromYamlPath(fn)
		if err != nil {
			if err == ErrInvalidPart {
				err = errOriginNotFound
			}
			return err
		}
	}

	// TODO: verify cache files here

	baseDir := filepath.Dir(filepath.Dir(fn))
	err = generatePolicy(m, baseDir)
	if err != nil {
		return err
	}

	return err
}

// RegenerateAllPolicy will re-generate all policy that needs re-generating
func RegenerateAllPolicy(force bool) error {
	installed, err := NewMetaLocalRepository().Installed()
	if err != nil {
		return err
	}

	for _, p := range installed {
		part, ok := p.(*SnapPart)
		if !ok {
			continue
		}
		basedir := part.basedir
		yFn := filepath.Join(basedir, "meta", "snap.yaml")

		// FIXME: use ErrPolicyNeedsRegenerating here to check if
		//        re-generation is needed
		if err := CompareGeneratePolicyFromFile(yFn); err == nil {
			continue
		}

		// re-generate!
		logger.Noticef("re-generating security policy for %s", yFn)
		if err := GeneratePolicyFromFile(yFn, force); err != nil {
			return err
		}
	}

	return nil
}
