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

package systemd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/ubuntu-core/snappy/dirs"
	"github.com/ubuntu-core/snappy/helpers"
)

type testreporter struct {
	msgs []string
}

func (tr *testreporter) Notify(msg string) {
	tr.msgs = append(tr.msgs, msg)
}

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

// systemd's testsuite
type SystemdTestSuite struct {
	i      int
	argses [][]string
	errors []error
	outs   [][]byte

	j     int
	jsvcs [][]string
	jouts [][]byte
	jerrs []error

	rep *testreporter
}

var _ = Suite(&SystemdTestSuite{})

func (s *SystemdTestSuite) SetUpTest(c *C) {
	// force UTC timezone, for reproducible timestamps
	os.Setenv("TZ", "")

	SystemctlCmd = s.myRun
	s.i = 0
	s.argses = nil
	s.errors = nil
	s.outs = nil

	JournalctlCmd = s.myJctl
	s.j = 0
	s.jsvcs = nil
	s.jouts = nil
	s.jerrs = nil

	s.rep = new(testreporter)
}

func (s *SystemdTestSuite) TearDownTest(c *C) {
	SystemctlCmd = run
	JournalctlCmd = jctl
}

func (s *SystemdTestSuite) myRun(args ...string) (out []byte, err error) {
	s.argses = append(s.argses, args)
	if s.i < len(s.outs) {
		out = s.outs[s.i]
	}
	if s.i < len(s.errors) {
		err = s.errors[s.i]
	}
	s.i++
	return out, err
}

func (s *SystemdTestSuite) myJctl(svcs []string) (out []byte, err error) {
	s.jsvcs = append(s.jsvcs, svcs)

	if s.j < len(s.jouts) {
		out = s.jouts[s.j]
	}
	if s.j < len(s.jerrs) {
		err = s.jerrs[s.j]
	}
	s.j++

	return out, err
}

func (s *SystemdTestSuite) errorRun(args ...string) (out []byte, err error) {
	return nil, &Error{cmd: args, exitCode: 1, msg: []byte("error on error")}
}

func (s *SystemdTestSuite) TestDaemonReload(c *C) {
	err := New("", s.rep).DaemonReload()
	c.Assert(err, IsNil)
	c.Assert(s.argses, DeepEquals, [][]string{{"daemon-reload"}})
}

func (s *SystemdTestSuite) TestStart(c *C) {
	err := New("", s.rep).Start("foo")
	c.Assert(err, IsNil)
	c.Check(s.argses, DeepEquals, [][]string{{"start", "foo"}})
}

func (s *SystemdTestSuite) TestStop(c *C) {
	s.outs = [][]byte{
		nil, // for the "stop" itself
		[]byte("ActiveState=whatever\n"),
		[]byte("ActiveState=active\n"),
		[]byte("ActiveState=inactive\n"),
	}
	s.errors = []error{nil, nil, nil, nil, &Timeout{}}
	err := New("", s.rep).Stop("foo", time.Millisecond)
	c.Assert(err, IsNil)
	c.Assert(s.argses, HasLen, 4)
	c.Check(s.argses[0], DeepEquals, []string{"stop", "foo"})
	c.Check(s.argses[1], DeepEquals, []string{"show", "--property=ActiveState", "foo"})
	c.Check(s.argses[1], DeepEquals, s.argses[2])
	c.Check(s.argses[1], DeepEquals, s.argses[3])
}

func (s *SystemdTestSuite) TestStatus(c *C) {
	s.outs = [][]byte{
		[]byte("Id=Thing\nLoadState=LoadState\nActiveState=ActiveState\nSubState=SubState\nUnitFileState=UnitFileState\n"),
	}
	s.errors = []error{nil}
	out, err := New("", s.rep).Status("foo")
	c.Assert(err, IsNil)
	c.Check(out, Equals, "UnitFileState; LoadState; ActiveState (SubState)")
}

func (s *SystemdTestSuite) TestStatusObj(c *C) {
	s.outs = [][]byte{
		[]byte("Id=Thing\nLoadState=LoadState\nActiveState=ActiveState\nSubState=SubState\nUnitFileState=UnitFileState\n"),
	}
	s.errors = []error{nil}
	out, err := New("", s.rep).ServiceStatus("foo")
	c.Assert(err, IsNil)
	c.Check(out, DeepEquals, &ServiceStatus{
		ServiceFileName: "foo",
		LoadState:       "LoadState",
		ActiveState:     "ActiveState",
		SubState:        "SubState",
		UnitFileState:   "UnitFileState",
	})
}

func (s *SystemdTestSuite) TestStopTimeout(c *C) {
	oldSteps := stopSteps
	oldDelay := stopDelay
	stopSteps = 2
	stopDelay = time.Millisecond
	defer func() {
		stopSteps = oldSteps
		stopDelay = oldDelay
	}()

	err := New("", s.rep).Stop("foo", 10*time.Millisecond)
	c.Assert(err, FitsTypeOf, &Timeout{})
	c.Check(s.rep.msgs[0], Equals, "Waiting for foo to stop.")
}

func (s *SystemdTestSuite) TestDisable(c *C) {
	err := New("xyzzy", s.rep).Disable("foo")
	c.Assert(err, IsNil)
	c.Check(s.argses, DeepEquals, [][]string{{"--root", "xyzzy", "disable", "foo"}})
}

func (s *SystemdTestSuite) TestEnable(c *C) {
	sysd := New("xyzzy", s.rep)
	sysd.(*systemd).rootDir = c.MkDir()
	err := os.MkdirAll(filepath.Join(sysd.(*systemd).rootDir, "/etc/systemd/system/multi-user.target.wants"), 0755)
	c.Assert(err, IsNil)

	err = sysd.Enable("foo")
	c.Assert(err, IsNil)

	// check symlink
	enableLink := filepath.Join(sysd.(*systemd).rootDir, "/etc/systemd/system/multi-user.target.wants/foo")
	target, err := os.Readlink(enableLink)
	c.Assert(err, IsNil)
	c.Assert(target, Equals, "/etc/systemd/system/foo")
}

const expectedServiceFmt = `[Unit]
Description=descr
%s
X-Snappy=yes

[Service]
ExecStart=/usr/bin/ubuntu-core-launcher app%[2]s aa-profile /apps/app%[2]s/1.0/bin/start
Restart=on-failure
WorkingDirectory=/apps/app%[2]s/1.0/
Environment="SNAP_APP=app_service_1.0" "TMPDIR=/tmp/snaps/app%[2]s/1.0/tmp" "TEMPDIR=/tmp/snaps/app%[2]s/1.0/tmp" "SNAP_APP_PATH=/apps/app%[2]s/1.0/" "SNAP_APP_DATA_PATH=/var/lib/apps/app%[2]s/1.0/" "SNAP_APP_TMPDIR=/tmp/snaps/app%[2]s/1.0/tmp" "SNAP_NAME=app" "SNAP_VERSION=1.0" "SNAP_ORIGIN=%[3]s" "SNAP_FULLNAME=app%[2]s" "SNAP_ARCH=%[5]s" "SNAP_APP_USER_DATA_PATH=%%h/apps/app%[2]s/1.0/" "SNAPP_APP_PATH=/apps/app%[2]s/1.0/" "SNAPP_APP_DATA_PATH=/var/lib/apps/app%[2]s/1.0/" "SNAPP_APP_TMPDIR=/tmp/snaps/app%[2]s/1.0/tmp" "SNAPPY_APP_ARCH=%[5]s" "SNAPP_APP_USER_DATA_PATH=%%h/apps/app%[2]s/1.0/"
ExecStop=/usr/bin/ubuntu-core-launcher app%[2]s aa-profile /apps/app%[2]s/1.0/bin/stop
ExecStopPost=/usr/bin/ubuntu-core-launcher app%[2]s aa-profile /apps/app%[2]s/1.0/bin/stop --post
TimeoutStopSec=10
%[4]s

[Install]
WantedBy=multi-user.target
`

var (
	expectedAppService  = fmt.Sprintf(expectedServiceFmt, "After=ubuntu-snappy.frameworks.target\nRequires=ubuntu-snappy.frameworks.target", ".mvo", "mvo", "\n", helpers.UbuntuArchitecture())
	expectedFmkService  = fmt.Sprintf(expectedServiceFmt, "Before=ubuntu-snappy.frameworks.target\nAfter=ubuntu-snappy.frameworks-pre.target\nRequires=ubuntu-snappy.frameworks-pre.target", "", "", "\n", helpers.UbuntuArchitecture())
	expectedDbusService = fmt.Sprintf(expectedServiceFmt, "After=ubuntu-snappy.frameworks.target\nRequires=ubuntu-snappy.frameworks.target", ".mvo", "mvo", "BusName=foo.bar.baz\nType=dbus", helpers.UbuntuArchitecture())

	// things that need network
	expectedNetAppService = fmt.Sprintf(expectedServiceFmt, "After=ubuntu-snappy.frameworks.target\nRequires=ubuntu-snappy.frameworks.target\nAfter=snappy-wait4network.service\nRequires=snappy-wait4network.service", ".mvo", "mvo", "\n", helpers.UbuntuArchitecture())
	expectedNetFmkService = fmt.Sprintf(expectedServiceFmt, "Before=ubuntu-snappy.frameworks.target\nAfter=ubuntu-snappy.frameworks-pre.target\nRequires=ubuntu-snappy.frameworks-pre.target\nAfter=snappy-wait4network.service\nRequires=snappy-wait4network.service", "", "", "\n", helpers.UbuntuArchitecture())
)

func (s *SystemdTestSuite) TestGenAppServiceFile(c *C) {

	desc := &ServiceDescription{
		AppName:     "app",
		ServiceName: "service",
		Version:     "1.0",
		Description: "descr",
		AppPath:     "/apps/app.mvo/1.0/",
		Start:       "bin/start",
		Stop:        "bin/stop",
		PostStop:    "bin/stop --post",
		StopTimeout: time.Duration(10 * time.Second),
		AaProfile:   "aa-profile",
		UdevAppName: "app.mvo",
	}

	c.Check(New("", nil).GenServiceFile(desc), Equals, expectedAppService)
}

func (s *SystemdTestSuite) TestGenNetAppServiceFile(c *C) {

	desc := &ServiceDescription{
		AppName:     "app",
		ServiceName: "service",
		Version:     "1.0",
		Description: "descr",
		AppPath:     "/apps/app.mvo/1.0/",
		Start:       "bin/start",
		Stop:        "bin/stop",
		PostStop:    "bin/stop --post",
		StopTimeout: time.Duration(10 * time.Second),
		AaProfile:   "aa-profile",
		IsNetworked: true,
		UdevAppName: "app.mvo",
	}

	c.Check(New("", nil).GenServiceFile(desc), Equals, expectedNetAppService)
}

func (s *SystemdTestSuite) TestGenFmkServiceFile(c *C) {

	desc := &ServiceDescription{
		AppName:     "app",
		ServiceName: "service",
		Version:     "1.0",
		Description: "descr",
		AppPath:     "/apps/app/1.0/",
		Start:       "bin/start",
		Stop:        "bin/stop",
		PostStop:    "bin/stop --post",
		StopTimeout: time.Duration(10 * time.Second),
		AaProfile:   "aa-profile",
		IsFramework: true,
		UdevAppName: "app",
	}

	c.Check(New("", nil).GenServiceFile(desc), Equals, expectedFmkService)
}

func (s *SystemdTestSuite) TestGenNetFmkServiceFile(c *C) {

	desc := &ServiceDescription{
		AppName:     "app",
		ServiceName: "service",
		Version:     "1.0",
		Description: "descr",
		AppPath:     "/apps/app/1.0/",
		Start:       "bin/start",
		Stop:        "bin/stop",
		PostStop:    "bin/stop --post",
		StopTimeout: time.Duration(10 * time.Second),
		AaProfile:   "aa-profile",
		IsNetworked: true,
		IsFramework: true,
		UdevAppName: "app",
	}

	c.Check(New("", nil).GenServiceFile(desc), Equals, expectedNetFmkService)
}

func (s *SystemdTestSuite) TestGenServiceFileWithBusName(c *C) {

	desc := &ServiceDescription{
		AppName:     "app",
		ServiceName: "service",
		Version:     "1.0",
		Description: "descr",
		AppPath:     "/apps/app.mvo/1.0/",
		Start:       "bin/start",
		Stop:        "bin/stop",
		PostStop:    "bin/stop --post",
		StopTimeout: time.Duration(10 * time.Second),
		AaProfile:   "aa-profile",
		BusName:     "foo.bar.baz",
		UdevAppName: "app.mvo",
	}

	generated := New("", nil).GenServiceFile(desc)
	c.Assert(generated, Equals, expectedDbusService)
}

func (s *SystemdTestSuite) TestRestart(c *C) {
	s.outs = [][]byte{
		nil, // for the "stop" itself
		[]byte("ActiveState=inactive\n"),
		nil, // for the "start"
	}
	s.errors = []error{nil, nil, nil, nil, &Timeout{}}
	err := New("", s.rep).Restart("foo", time.Millisecond)
	c.Assert(err, IsNil)
	c.Check(s.argses, HasLen, 3)
	c.Check(s.argses[0], DeepEquals, []string{"stop", "foo"})
	c.Check(s.argses[1], DeepEquals, []string{"show", "--property=ActiveState", "foo"})
	c.Check(s.argses[2], DeepEquals, []string{"start", "foo"})
}

func (s *SystemdTestSuite) TestKill(c *C) {
	c.Assert(New("", s.rep).Kill("foo", "HUP"), IsNil)
	c.Check(s.argses, DeepEquals, [][]string{{"kill", "foo", "-s", "HUP"}})
}

func (s *SystemdTestSuite) TestIsTimeout(c *C) {
	c.Check(IsTimeout(os.ErrInvalid), Equals, false)
	c.Check(IsTimeout(&Timeout{}), Equals, true)
}

func (s *SystemdTestSuite) TestLogErrJctl(c *C) {
	s.jerrs = []error{&Timeout{}}

	logs, err := New("", s.rep).Logs([]string{"foo"})
	c.Check(err, NotNil)
	c.Check(logs, IsNil)
	c.Check(s.jsvcs, DeepEquals, [][]string{{"foo"}})
	c.Check(s.j, Equals, 1)
}

func (s *SystemdTestSuite) TestLogErrJSON(c *C) {
	s.jouts = [][]byte{[]byte("this is not valid json.")}

	logs, err := New("", s.rep).Logs([]string{"foo"})
	c.Check(err, NotNil)
	c.Check(logs, IsNil)
	c.Check(s.jsvcs, DeepEquals, [][]string{{"foo"}})
	c.Check(s.j, Equals, 1)
}

func (s *SystemdTestSuite) TestLogs(c *C) {
	s.jouts = [][]byte{[]byte(`{"a": 1}
{"a": 2}
`)}

	logs, err := New("", s.rep).Logs([]string{"foo"})
	c.Check(err, IsNil)
	c.Check(logs, DeepEquals, []Log{{"a": 1.}, {"a": 2.}})
	c.Check(s.jsvcs, DeepEquals, [][]string{{"foo"}})
	c.Check(s.j, Equals, 1)
}

func (s *SystemdTestSuite) TestLogString(c *C) {
	c.Check(Log{}.String(), Equals, "-(no timestamp!)- - -")
	c.Check(Log{
		"__REALTIME_TIMESTAMP": 42,
	}.String(), Equals, "-(timestamp not a string: 42)- - -")
	c.Check(Log{
		"__REALTIME_TIMESTAMP": "what",
	}.String(), Equals, "-(timestamp not a decimal number: \"what\")- - -")
	c.Check(Log{
		"__REALTIME_TIMESTAMP": "0",
		"MESSAGE":              "hi",
	}.String(), Equals, "1970-01-01T00:00:00.000000Z - hi")
	c.Check(Log{
		"__REALTIME_TIMESTAMP": "42",
		"MESSAGE":              "hi",
		"SYSLOG_IDENTIFIER":    "me",
	}.String(), Equals, "1970-01-01T00:00:00.000042Z me hi")

}

func (s *SystemdTestSuite) TestMountUnitPath(c *C) {
	c.Assert(MountUnitPath("/apps/hello.origin/1.1", "mount"), Equals, filepath.Join(dirs.SnapServicesDir, "apps-hello.origin-1.1.mount"))
}

func (s *SystemdTestSuite) TestWriteMountUnit(c *C) {
	mountUnitName, err := New("", nil).WriteMountUnitFile("foo.origin", "/var/lib/snappy/snaps/foo.origin_1.0.snap", "/apps/foo.origin/1.0")
	c.Assert(err, IsNil)

	mount, err := ioutil.ReadFile(filepath.Join(dirs.SnapServicesDir, mountUnitName))
	c.Assert(err, IsNil)
	c.Assert(string(mount), Equals, `[Unit]
Description=Snapfs mount unit for foo.origin

[Mount]
What=/var/lib/snappy/snaps/foo.origin_1.0.snap
Where=/apps/foo.origin/1.0
`)
}

func (s *SystemdTestSuite) TestWriteAutoMountUnit(c *C) {
	mountUnitName, err := New("", nil).WriteAutoMountUnitFile("foo.origin", "/apps/foo.origin/1.0")
	c.Assert(err, IsNil)

	automount, err := ioutil.ReadFile(filepath.Join(dirs.SnapServicesDir, mountUnitName))
	c.Assert(err, IsNil)
	c.Assert(string(automount), Equals, `[Unit]
Description=Snapfs automount unit for foo.origin

[Automount]
Where=/apps/foo.origin/1.0
TimeoutIdleSec=30

[Install]
WantedBy=multi-user.target
`)
}
