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

package snappy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	. "gopkg.in/check.v1"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/helpers"
	"github.com/ubuntu-core/snappy/logger"
	"github.com/ubuntu-core/snappy/security"
	"github.com/ubuntu-core/snappy/snap"
	"github.com/ubuntu-core/snappy/snap/app"
)

type SecurityTestSuite struct {
	tempDir               string
	buildDir              string
	m                     *snapYaml
	scFilterGenCall       []string
	scFilterGenCallReturn []byte

	loadAppArmorPolicyCalled bool
}

var _ = Suite(&SecurityTestSuite{})

func (a *SecurityTestSuite) SetUpTest(c *C) {
	a.buildDir = c.MkDir()
	a.tempDir = c.MkDir()
	os.MkdirAll(filepath.Join(a.buildDir, "meta"), 0755)

	// set global sandbox
	dirs.SetRootDir(c.MkDir())

	a.m = &snapYaml{
		Name:    "foo",
		Version: "1.0",
	}

	// and mock some stuff
	a.loadAppArmorPolicyCalled = false
	loadAppArmorPolicy = func(fn string) ([]byte, error) {
		a.loadAppArmorPolicyCalled = true
		return nil, nil
	}
	runUdevAdm = func(args ...string) error {
		return nil
	}
}

func (a *SecurityTestSuite) TearDownTest(c *C) {
	dirs.SetRootDir("/")
}

func ensureFileContentMatches(c *C, fn, expectedContent string) {
	content, err := ioutil.ReadFile(fn)
	c.Assert(err, IsNil)
	c.Assert(string(content), Equals, expectedContent)
}

// mocks
func defaultPolicyVendor() string {
	return "ubuntu-core"
}
func defaultPolicyVersion() string {
	return "16.04"
}

func makeMockSecurityEnv(c *C) {
	makeMockApparmorTemplate(c, "default", []byte(""))
	makeMockSeccompTemplate(c, "default", []byte(""))
	makeMockApparmorCap(c, "network-client", []byte(``))
	makeMockSeccompCap(c, "network-client", []byte(``))
}

func makeMockApparmorTemplate(c *C, templateName string, content []byte) {
	mockTemplate := filepath.Join(security.PolicyTypeAppArmor.PolicyDir(), "templates", defaultPolicyVendor(), defaultPolicyVersion(), templateName)
	err := os.MkdirAll(filepath.Dir(mockTemplate), 0755)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(mockTemplate, content, 0644)
	c.Assert(err, IsNil)
}

func makeMockApparmorCap(c *C, capname string, content []byte) {
	mockPG := filepath.Join(security.PolicyTypeAppArmor.PolicyDir(), "policygroups", defaultPolicyVendor(), defaultPolicyVersion(), capname)
	err := os.MkdirAll(filepath.Dir(mockPG), 0755)
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(mockPG, []byte(content), 0644)
	c.Assert(err, IsNil)
}

func makeMockSeccompTemplate(c *C, templateName string, content []byte) {
	mockTemplate := filepath.Join(security.PolicyTypeSeccomp.PolicyDir(), "templates", defaultPolicyVendor(), defaultPolicyVersion(), templateName)
	err := os.MkdirAll(filepath.Dir(mockTemplate), 0755)
	c.Assert(err, IsNil)
	err = ioutil.WriteFile(mockTemplate, content, 0644)
	c.Assert(err, IsNil)
}

func makeMockSeccompCap(c *C, capname string, content []byte) {
	mockPG := filepath.Join(security.PolicyTypeSeccomp.PolicyDir(), "policygroups", defaultPolicyVendor(), defaultPolicyVersion(), capname)
	err := os.MkdirAll(filepath.Dir(mockPG), 0755)
	c.Assert(err, IsNil)

	err = ioutil.WriteFile(mockPG, []byte(content), 0644)
	c.Assert(err, IsNil)
}

func (a *SecurityTestSuite) TestSecurityFindWhitespacePrefix(c *C) {
	t := `  ###POLICYGROUPS###`
	c.Assert(findWhitespacePrefix(t, "###POLICYGROUPS###"), Equals, "  ")

	t = `not there`
	c.Assert(findWhitespacePrefix(t, "###POLICYGROUPS###"), Equals, "")

	t = `not there`
	c.Assert(findWhitespacePrefix(t, "###POLICYGROUPS###"), Equals, "")
}

func (a *SecurityTestSuite) TestSecurityFindWhitespacePrefixNeedsQuoting(c *C) {
	s := `I need quoting: [`
	t := ``
	c.Assert(findWhitespacePrefix(s, t), Equals, t)
}

// FIXME: need additional test for frameworkPolicy
func (a *SecurityTestSuite) TestSecurityFindTemplateApparmor(c *C) {
	makeMockApparmorTemplate(c, "mock-template", []byte(`something`))

	t, err := security.PolicyTypeAppArmor.FindTemplate("mock-template")
	c.Assert(err, IsNil)
	c.Assert(t, Matches, "something")
}

func (a *SecurityTestSuite) TestSecurityFindTemplateApparmorNotFound(c *C) {
	_, err := security.PolicyTypeAppArmor.FindTemplate("not-available-templ")
	c.Assert(err, DeepEquals, &security.ErrPolicyNotFound{"template", &security.PolicyTypeAppArmor, "not-available-templ"})
}

// FIXME: need additional test for frameworkPolicy
func (a *SecurityTestSuite) TestSecurityFindCaps(c *C) {
	for _, f := range []string{"cap1", "cap2"} {
		makeMockApparmorCap(c, f, []byte(f))
	}

	cap, err := security.PolicyTypeAppArmor.FindCaps([]string{"cap1", "cap2"}, "mock-template")
	c.Assert(err, IsNil)
	c.Assert(cap, DeepEquals, []string{"cap1", "cap2"})
}

func (a *SecurityTestSuite) TestSecurityFindCapsMultipleErrorHandling(c *C) {
	makeMockApparmorCap(c, "existing-cap", []byte("something"))

	_, err := security.PolicyTypeAppArmor.FindCaps([]string{"existing-cap", "not-existing-cap"}, "mock-template")
	c.Check(err, ErrorMatches, "could not find specified cap: not-existing-cap.*")

	_, err = security.PolicyTypeAppArmor.FindCaps([]string{"not-existing-cap", "existing-cap"}, "mock-template")
	c.Check(err, ErrorMatches, "could not find specified cap: not-existing-cap.*")

	_, err = security.PolicyTypeAppArmor.FindCaps([]string{"existing-cap"}, "mock-template")
	c.Check(err, IsNil)
}

func (a *SecurityTestSuite) TestSecurityGenAppArmorPathRuleSimple(c *C) {
	pr, err := genAppArmorPathRule("/some/path", "rk")
	c.Assert(err, IsNil)
	c.Assert(pr, Equals, "/some/path rk,\n")
}

func (a *SecurityTestSuite) TestSecurityGenAppArmorPathRuleDir(c *C) {
	pr, err := genAppArmorPathRule("/some/path/", "rk")
	c.Assert(err, IsNil)
	c.Assert(pr, Equals, `/some/path/ rk,
/some/path/** rk,
`)
}

func (a *SecurityTestSuite) TestSecurityGenAppArmorPathRuleDirGlob(c *C) {
	pr, err := genAppArmorPathRule("/some/path/**", "rk")
	c.Assert(err, IsNil)
	c.Assert(pr, Equals, `/some/path/ rk,
/some/path/** rk,
`)
}

func (a *SecurityTestSuite) TestSecurityGenAppArmorPathRuleHome(c *C) {
	pr, err := genAppArmorPathRule("/home/something", "rk")
	c.Assert(err, IsNil)
	c.Assert(pr, Equals, "owner /home/something rk,\n")
}

func (a *SecurityTestSuite) TestSecurityGenAppArmorPathRuleError(c *C) {
	_, err := genAppArmorPathRule("some/path", "rk")
	c.Assert(err, Equals, errPolicyGen)
}

var mockApparmorTemplate = []byte(`
# Description: Allows unrestricted access to the system
# Usage: reserved

# vim:syntax=apparmor

#include <tunables/global>

# Define vars with unconfined since autopilot rules may reference them
###VAR###

# v2 compatible wildly permissive profile
###PROFILEATTACH### (attach_disconnected) {
  capability,
  network,
  / rwkl,
  /** rwlkm,
  # Ubuntu Core is a minimal system so don't use 'pix' here. There are few
  # profiles to transition to, and those that exist either won't work right
  # anyway (eg, ubuntu-core-launcher) or would need to be modified to work
  # with snaps (dhclient).
  /** ix,

  mount,
  remount,

  ###ABSTRACTIONS###

  ###POLICYGROUPS###

  ###READS###

  ###WRITES###
}`)

var expectedGeneratedAaProfile = `
# Description: Allows unrestricted access to the system
# Usage: reserved

# vim:syntax=apparmor

#include <tunables/global>

# Define vars with unconfined since autopilot rules may reference them
# Specified profile variables
@{APP_APPNAME}=""
@{APP_ID_DBUS}=""
@{APP_PKGNAME_DBUS}="foo"
@{APP_PKGNAME}="foo"
@{APP_VERSION}="1.0"
@{INSTALL_DIR}="{/snaps,/gadget}"
# Deprecated:
@{CLICK_DIR}="{/snaps,/gadget}"

# v2 compatible wildly permissive profile
profile "" (attach_disconnected) {
  capability,
  network,
  / rwkl,
  /** rwlkm,
  # Ubuntu Core is a minimal system so don't use 'pix' here. There are few
  # profiles to transition to, and those that exist either won't work right
  # anyway (eg, ubuntu-core-launcher) or would need to be modified to work
  # with snaps (dhclient).
  /** ix,

  mount,
  remount,

  # No abstractions specified

  # Rules specified via caps (policy groups)
  capito

  # No read paths specified

  # No write paths specified
}`

func (a *SecurityTestSuite) TestSecurityGenAppArmorTemplatePolicy(c *C) {
	makeMockApparmorTemplate(c, "mock-template", mockApparmorTemplate)
	makeMockApparmorCap(c, "cap1", []byte(`capito`))

	m := &snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	appid := &security.AppID{
		Pkgname: "foo",
		Version: "1.0",
	}
	template := "mock-template"
	caps := []string{"cap1"}
	overrides := &security.OverrideDefinition{}
	p, err := getAppArmorTemplatedPolicy(m, appid, template, caps, overrides)
	c.Check(err, IsNil)
	c.Check(p, Equals, expectedGeneratedAaProfile)
}

var mockSeccompTemplate = []byte(`
# Description: Allows access to app-specific directories and basic runtime
# Usage: common
#

# Dangerous syscalls that we don't ever want to allow

# kexec
deny kexec_load

# fine
alarm
`)

var expectedGeneratedSeccompProfile = `
# Description: Allows access to app-specific directories and basic runtime
# Usage: common
#

# Dangerous syscalls that we don't ever want to allow

# kexec
# EXPLICITLY DENIED: kexec_load

# fine
alarm

#cap1
capino`

func (a *SecurityTestSuite) TestSecurityGenSeccompTemplatedPolicy(c *C) {
	makeMockSeccompTemplate(c, "mock-template", mockSeccompTemplate)
	makeMockSeccompCap(c, "cap1", []byte("#cap1\ncapino\n"))

	m := &snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	appid := &security.AppID{
		Pkgname: "foo",
		Version: "1.0",
	}
	template := "mock-template"
	caps := []string{"cap1"}
	overrides := &security.OverrideDefinition{}
	p, err := getSeccompTemplatedPolicy(m, appid, template, caps, overrides)
	c.Check(err, IsNil)
	c.Check(p, Equals, expectedGeneratedSeccompProfile)
}

var aaCustomPolicy = `
# Description: Some custom aa policy
# Usage: reserved

# vim:syntax=apparmor

#include <tunables/global>

# Define vars with unconfined since autopilot rules may reference them
###VAR###

# v2 compatible wildly permissive profile
###PROFILEATTACH### (attach_disconnected) {
  capability,
}
`
var expectedAaCustomPolicy = `
# Description: Some custom aa policy
# Usage: reserved

# vim:syntax=apparmor

#include <tunables/global>

# Define vars with unconfined since autopilot rules may reference them
# Specified profile variables
@{APP_APPNAME}=""
@{APP_ID_DBUS}="foo_5fbar_5f1_2e0"
@{APP_PKGNAME_DBUS}="foo"
@{APP_PKGNAME}="foo"
@{APP_VERSION}="1.0"
@{INSTALL_DIR}="{/snaps,/gadget}"
# Deprecated:
@{CLICK_DIR}="{/snaps,/gadget}"

# v2 compatible wildly permissive profile
profile "foo_bar_1.0" (attach_disconnected) {
  capability,

# No read paths specified
# No write paths specified
# No abstractions specified
}
`

func (a *SecurityTestSuite) TestSecurityGetApparmorCustomPolicy(c *C) {
	m := &snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	appid := &security.AppID{
		AppID:   "foo_bar_1.0",
		Pkgname: "foo",
		Version: "1.0",
	}
	customPolicy := filepath.Join(c.MkDir(), "foo")
	err := ioutil.WriteFile(customPolicy, []byte(aaCustomPolicy), 0644)
	c.Assert(err, IsNil)

	p, err := getAppArmorCustomPolicy(m, appid, customPolicy, nil)
	c.Check(err, IsNil)
	c.Check(p, Equals, expectedAaCustomPolicy)
}

func (a *SecurityTestSuite) TestSecurityGetSeccompCustomPolicy(c *C) {
	// yes, getSeccompCustomPolicy does not care for snapYaml or appid
	m := &snap.Info{}
	appid := &security.AppID{}

	customPolicy := filepath.Join(c.MkDir(), "foo")
	err := ioutil.WriteFile(customPolicy, []byte(`canary`), 0644)
	c.Assert(err, IsNil)

	p, err := getSeccompCustomPolicy(m, appid, customPolicy)
	c.Check(err, IsNil)
	c.Check(p, Equals, `canary`)
}

func (a *SecurityTestSuite) TestSecurityMergeApparmorSecurityOverridesNilDoesNotCrash(c *C) {
	sd := &security.Definitions{}
	sd.MergeAppArmorSecurityOverrides(nil)
	c.Assert(sd, DeepEquals, &security.Definitions{})
}

func (a *SecurityTestSuite) TestSecurityMergeApparmorSecurityOverridesTrivial(c *C) {
	sd := &security.Definitions{}
	hwaccessOverrides := &security.OverrideDefinition{}
	sd.MergeAppArmorSecurityOverrides(hwaccessOverrides)

	c.Assert(sd, DeepEquals, &security.Definitions{
		SecurityOverride: hwaccessOverrides,
	})
}

func (a *SecurityTestSuite) TestSecurityMergeApparmorSecurityOverridesOverrides(c *C) {
	sd := &security.Definitions{}
	hwaccessOverrides := &security.OverrideDefinition{
		ReadPaths:  []string{"read1"},
		WritePaths: []string{"write1"},
	}
	sd.MergeAppArmorSecurityOverrides(hwaccessOverrides)

	c.Assert(sd, DeepEquals, &security.Definitions{
		SecurityOverride: hwaccessOverrides,
	})
}

func (a *SecurityTestSuite) TestSecurityMergeApparmorSecurityOverridesMerges(c *C) {
	sd := &security.Definitions{
		SecurityOverride: &security.OverrideDefinition{
			ReadPaths: []string{"orig1"},
		},
	}
	hwaccessOverrides := &security.OverrideDefinition{
		ReadPaths:  []string{"read1"},
		WritePaths: []string{"write1"},
	}
	sd.MergeAppArmorSecurityOverrides(hwaccessOverrides)

	c.Assert(sd, DeepEquals, &security.Definitions{
		SecurityOverride: &security.OverrideDefinition{
			ReadPaths:  []string{"orig1", "read1"},
			WritePaths: []string{"write1"},
		},
	})
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyForServiceBinaryEmpty(c *C) {
	makeMockApparmorTemplate(c, "default", []byte(`# apparmor
###POLICYGROUPS###
`))
	makeMockApparmorCap(c, "network-client", []byte(`
aa-network-client`))
	makeMockSeccompTemplate(c, "default", []byte(`write`))
	makeMockSeccompCap(c, "network-client", []byte(`
sc-network-client
`))

	// empty SecurityDefinition means "network-client" cap
	sd := &security.Definitions{}
	m := &snap.Info{
		Name:    "pkg",
		Version: "1.0",
	}

	// generate the apparmor profile
	err := generatePolicyForServiceBinary(sd, m, "binary", "/snaps/app.origin/1.0")
	c.Assert(err, IsNil)

	// ensure the apparmor policy got loaded
	c.Assert(a.loadAppArmorPolicyCalled, Equals, true)

	aaProfile := filepath.Join(dirs.SnapAppArmorDir, "pkg.origin_binary_1.0")
	ensureFileContentMatches(c, aaProfile, `# apparmor
# Rules specified via caps (policy groups)

aa-network-client
`)
	scProfile := filepath.Join(dirs.SnapSeccompDir, "pkg.origin_binary_1.0")
	ensureFileContentMatches(c, scProfile, `write

sc-network-client`)

}

var mockSecuritySnapYaml = `
name: hello-world
vendor: someone
version: 1.0
apps:
 binary1:
   uses: [binary1]
 service1:
   uses: [service1]
   daemon: forking
uses:
 binary1:
  type: migration-skill
  caps: []
 service1:
  type: migration-skill
  caps: []
`

func (a *SecurityTestSuite) TestSecurityGeneratePolicyFromFileSimple(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(`# some header
###POLICYGROUPS###
`))
	makeMockSeccompTemplate(c, "default", []byte(`
deny kexec
read
write
`))

	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)

	// the acutal thing that gets tested
	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// ensure the apparmor policy got loaded
	c.Assert(a.loadAppArmorPolicyCalled, Equals, true)

	// apparmor
	generatedProfileFn := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("hello-world.%s_binary1_1.0", testOrigin))
	ensureFileContentMatches(c, generatedProfileFn, `# some header
# No caps (policy groups) specified
`)
	// ... and seccomp
	generatedProfileFn = filepath.Join(dirs.SnapSeccompDir, fmt.Sprintf("hello-world.%s_binary1_1.0", testOrigin))
	ensureFileContentMatches(c, generatedProfileFn, `
# EXPLICITLY DENIED: kexec
read
write

`)
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyFileForConfig(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(`# some header
###POLICYGROUPS###
`))
	makeMockSeccompTemplate(c, "default", []byte(`
deny kexec
read
write
`))

	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)
	configHook := filepath.Join(filepath.Dir(mockSnapYamlFn), "hooks", "config")
	os.MkdirAll(filepath.Dir(configHook), 0755)
	err = ioutil.WriteFile(configHook, []byte("true"), 0755)
	c.Assert(err, IsNil)

	// generate config
	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// and for snappy-config
	generatedProfileFn := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("hello-world.%s_snappy-config_1.0", testOrigin))
	ensureFileContentMatches(c, generatedProfileFn, `# some header
# No caps (policy groups) specified
`)

}

func (a *SecurityTestSuite) TestSecurityCompareGeneratePolicyFromFileSimple(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(`# some header
###POLICYGROUPS###
`))
	makeMockSeccompTemplate(c, "default", []byte(`
deny kexec
read
write
`))
	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)

	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// nothing changed, compare is happy
	err = CompareGeneratePolicyFromFile(mockSnapYamlFn)
	c.Assert(err, IsNil)

	// now change the templates
	makeMockApparmorTemplate(c, "default", []byte(`# some different header
###POLICYGROUPS###
`))
	// ...and ensure that the difference is found
	err = CompareGeneratePolicyFromFile(mockSnapYamlFn)
	c.Assert(err, ErrorMatches, "policy differs.*")
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyFromFileHwAccess(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(`# some header
###POLICYGROUPS###
###READS###
###WRITES###
`))
	makeMockSeccompTemplate(c, "default", []byte(`
deny kexec
read
write
`))
	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)
	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// ensure that AddHWAccess does the right thing
	a.loadAppArmorPolicyCalled = false
	err = AddHWAccess("hello-world."+testOrigin, "/dev/kmesg")
	c.Assert(err, IsNil)

	// ensure the apparmor policy got loaded
	c.Check(a.loadAppArmorPolicyCalled, Equals, true)

	// apparmor got updated with the new read path
	generatedProfileFn := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("hello-world.%s_binary1_1.0", testOrigin))
	ensureFileContentMatches(c, generatedProfileFn, `# some header
# No caps (policy groups) specified
# Additional read-paths from security-override
/run/udev/data/ rk,
/run/udev/data/* rk,

# Additional write-paths from security-override
/dev/kmesg rwk,

`)
}

func (a *SecurityTestSuite) TestSecurityRegenerateAll(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(`# some header
###POLICYGROUPS###
`))
	makeMockSeccompTemplate(c, "default", []byte(`
deny kexec
read
write
`))
	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)

	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// now change the templates
	makeMockApparmorTemplate(c, "default", []byte(`# some different header
###POLICYGROUPS###
`))
	// ...and regenerate the templates
	err = RegenerateAllPolicy(false)
	c.Assert(err, IsNil)

	// ensure apparmor got updated with the new read path
	generatedProfileFn := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("hello-world.%s_binary1_1.0", testOrigin))
	ensureFileContentMatches(c, generatedProfileFn, `# some different header
# No caps (policy groups) specified
`)

}

func makeCustomAppArmorPolicy(c *C) string {
	content := []byte(`# custom apparmor policy
###VAR###

###PROFILEATTACH### (attach_disconnected) {
 stuff

}
`)
	fn := filepath.Join(c.MkDir(), "custom-aa-policy")
	err := ioutil.WriteFile(fn, content, 0644)
	c.Assert(err, IsNil)

	return fn
}

func (a *SecurityTestSuite) TestSecurityGenerateCustomPolicyAdditionalIsConsidered(c *C) {
	m := &snap.Info{
		Name:    "foo",
		Version: "1.0",
	}
	appid := &security.AppID{
		Pkgname: "foo",
		Version: "1.0",
	}
	fn := makeCustomAppArmorPolicy(c)

	content, err := getAppArmorCustomPolicy(m, appid, fn, nil)
	c.Assert(err, IsNil)
	c.Assert(content, Matches, `(?ms).*^# No read paths specified$`)
	c.Assert(content, Matches, `(?ms).*^# No write paths specified$`)
	c.Assert(content, Matches, `(?ms).*^# No abstractions specified$`)
}

var mockSecurityDeprecatedSnapYaml = `
name: hello-world
vendor: someone
version: 1.0
apps:
 binary1:
   uses: [binary1]
uses:
 binary1:
   type: migration-skill
   caps: []
`

var mockSecurityDeprecatedSnapYamlApparmor1 = `
   security-override:
    apparmor:
     read-path: [foo]
`
var mockSecurityDeprecatedSnapYamlApparmor2 = `
   security-override:
    apparmor: {}
`
var mockSecurityDeprecatedSnapYamlSeccomp1 = `
   security-override:
    seccomp: {}
`

var mockSecurityDeprecatedSnapYamlSeccomp2 = `
   security-override:
    seccomp:
     syscalls: [1]
`

type mockLogger struct {
	notice []string
	debug  []string
}

func (l *mockLogger) Notice(msg string) {
	l.notice = append(l.notice, msg)
}

func (l *mockLogger) Debug(msg string) {
	l.debug = append(l.debug, msg)
}

func (a *SecurityTestSuite) TestSecurityWarnsNot(c *C) {
	makeMockApparmorTemplate(c, "default", []byte(``))
	makeMockSeccompTemplate(c, "default", []byte(``))

	ml := &mockLogger{}
	logger.SetLogger(ml)

	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecurityDeprecatedSnapYaml)
	c.Assert(err, IsNil)

	err = GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	c.Assert(ml.notice, DeepEquals, []string(nil))
}

func (a *SecurityTestSuite) TestSecurityWarnsOnDeprecatedApparmor(c *C) {
	makeMockApparmorTemplate(c, "default", []byte(``))
	makeMockSeccompTemplate(c, "default", []byte(``))

	for _, s := range []string{mockSecurityDeprecatedSnapYamlApparmor1, mockSecurityDeprecatedSnapYamlApparmor2} {

		ml := &mockLogger{}
		logger.SetLogger(ml)

		mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecurityDeprecatedSnapYaml+s)
		c.Assert(err, IsNil)

		err = GeneratePolicyFromFile(mockSnapYamlFn, false)
		c.Assert(err, IsNil)

		c.Assert(ml.notice, DeepEquals, []string{"The security-override.apparmor key is no longer supported, please use use security-override directly"})
	}
}

func (a *SecurityTestSuite) TestSecurityWarnsOnDeprecatedSeccomp(c *C) {
	makeMockApparmorTemplate(c, "default", []byte(``))
	makeMockSeccompTemplate(c, "default", []byte(``))

	for _, s := range []string{mockSecurityDeprecatedSnapYamlSeccomp1, mockSecurityDeprecatedSnapYamlSeccomp2} {

		ml := &mockLogger{}
		logger.SetLogger(ml)

		mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecurityDeprecatedSnapYaml+s)
		c.Assert(err, IsNil)

		err = GeneratePolicyFromFile(mockSnapYamlFn, false)
		c.Assert(err, IsNil)

		c.Assert(ml.notice, DeepEquals, []string{"The security-override.seccomp key is no longer supported, please use use security-override directly"})
	}
}

func makeInstalledMockSnapSideloaded(c *C) string {
	mockSnapYamlFn, err := makeInstalledMockSnap(dirs.GlobalRootDir, mockSecuritySnapYaml)
	c.Assert(err, IsNil)
	// pretend its sideloaded
	basePath := regexp.MustCompile(`(.*)/hello-world.` + testOrigin).FindString(mockSnapYamlFn)
	oldPath := filepath.Join(basePath, "1.0")
	newPath := filepath.Join(basePath, "IsSideloadVer")
	err = os.Rename(oldPath, newPath)
	mockSnapYamlFn = filepath.Join(basePath, "IsSideloadVer", "meta", "snap.yaml")

	return mockSnapYamlFn
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyFromFileSideload(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(``))
	makeMockSeccompTemplate(c, "default", []byte(``))

	mockSnapYamlFn := makeInstalledMockSnapSideloaded(c)

	// the acutal thing that gets tested
	err := GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// ensure the apparmor policy got loaded
	c.Assert(a.loadAppArmorPolicyCalled, Equals, true)

	// apparmor
	generatedProfileFn := filepath.Join(dirs.SnapAppArmorDir, fmt.Sprintf("hello-world.%s_binary1_IsSideloadVer", testOrigin))
	c.Assert(helpers.FileExists(generatedProfileFn), Equals, true)

	// ... and seccomp
	generatedProfileFn = filepath.Join(dirs.SnapSeccompDir, fmt.Sprintf("hello-world.%s_binary1_IsSideloadVer", testOrigin))
	c.Assert(helpers.FileExists(generatedProfileFn), Equals, true)
}

func (a *SecurityTestSuite) TestSecurityCompareGeneratePolicyFromFileSideload(c *C) {
	// we need to create some fake data
	makeMockApparmorTemplate(c, "default", []byte(``))
	makeMockSeccompTemplate(c, "default", []byte(``))

	mockSnapYamlFn := makeInstalledMockSnapSideloaded(c)
	// generate policy
	err := GeneratePolicyFromFile(mockSnapYamlFn, false)
	c.Assert(err, IsNil)

	// nothing changed, ensure compare is happy even for sideloaded pkgs
	err = CompareGeneratePolicyFromFile(mockSnapYamlFn)
	c.Assert(err, IsNil)
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyForServiceBinaryFramework(c *C) {
	makeMockSecurityEnv(c)

	sd := &security.Definitions{}
	m := &snap.Info{
		Name:    "framework-name",
		Type:    "framework",
		Version: "1.0",
	}

	// generate the apparmor profile
	err := generatePolicyForServiceBinary(sd, m, "binary", "/snaps/framework-anem/1.0")
	c.Assert(err, IsNil)

	// ensure its available with the right names
	aaProfile := filepath.Join(dirs.SnapAppArmorDir, "framework-name_binary_1.0")
	ensureFileContentMatches(c, aaProfile, ``)
	scProfile := filepath.Join(dirs.SnapSeccompDir, "framework-name_binary_1.0")
	ensureFileContentMatches(c, scProfile, `
`)
}

func (a *SecurityTestSuite) TestSecurityGeneratePolicyForServiceBinaryErrors(c *C) {
	makeMockSecurityEnv(c)

	sd := &security.Definitions{}
	m := &snap.Info{
		Name:    "app",
		Version: "1.0",
	}

	// ensure invalid packages generate an error
	err := generatePolicyForServiceBinary(sd, m, "binary", "/snaps/app-no-origin/1.0")
	c.Assert(err, ErrorMatches, "can not get origin from path.*")
}

func (a *SecurityTestSuite) TestParseSnapYamlWithVersion(c *C) {
	testVersion := "1.0"
	dir := filepath.Join(a.tempDir, "foo", testVersion, "meta")
	os.MkdirAll(dir, 0755)
	y := filepath.Join(dir, "snap.yaml")
	ioutil.WriteFile(y, []byte(`
name: foo
version: 123456789
`), 0644)
	m, err := parseSnapYamlFileWithVersion(y)
	c.Assert(err, IsNil)
	c.Assert(m.Version, Equals, testVersion)
}

func (a *SecurityTestSuite) TestParseSnapYamlWithVersionSymlink(c *C) {
	testVersion := "1.0"
	verDir := filepath.Join(a.tempDir, "foo", testVersion)
	symDir := filepath.Join(a.tempDir, "foo", "current")
	os.MkdirAll(filepath.Join(verDir, "meta"), 0755)
	os.Symlink(verDir, symDir)
	y := filepath.Join(symDir, "meta", "snap.yaml")
	ioutil.WriteFile(y, []byte(`
name: foo
version: 123456789
`), 0644)
	m, err := parseSnapYamlFileWithVersion(y)
	c.Assert(err, IsNil)
	c.Assert(m.Version, Equals, testVersion)

}

func (a *SecurityTestSuite) TestFindSkillForAppEmpty(c *C) {
	app := &app.Yaml{}
	m := &snapYaml{}
	skill, err := findSkillForApp(m, app)
	c.Check(err, IsNil)
	c.Check(skill, IsNil)
}

func (a *SecurityTestSuite) TestFindSkillForAppTooMany(c *C) {
	app := &app.Yaml{
		UsesRef: []string{"one", "two"},
	}
	m := &snapYaml{}
	skill, err := findSkillForApp(m, app)
	c.Check(skill, IsNil)
	c.Check(err, ErrorMatches, "only a single skill is supported, 2 found")
}

func (a *SecurityTestSuite) TestFindSkillForAppNotFound(c *C) {
	app := &app.Yaml{
		UsesRef: []string{"not-there"},
	}
	m := &snapYaml{}
	skill, err := findSkillForApp(m, app)
	c.Check(skill, IsNil)
	c.Check(err, ErrorMatches, `can not find skill "not-there"`)
}

func (a *SecurityTestSuite) TestFindSkillFinds(c *C) {
	app := &app.Yaml{
		UsesRef: []string{"skill"},
	}
	m := &snapYaml{
		Uses: map[string]*usesYaml{
			"skill": &usesYaml{Type: "some-type"},
		},
	}
	skill, err := findSkillForApp(m, app)
	c.Check(err, IsNil)
	c.Check(skill.Type, Equals, "some-type")
}
