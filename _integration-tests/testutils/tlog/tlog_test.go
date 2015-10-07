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

package tlog

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

// Different suite so that we don't initialize fields
// of the subject and can do assertions on the default
// values
type logTestDefaultValuesSuite struct {
}

var _ = check.Suite(&logTestDefaultValuesSuite{})

func (s *logTestDefaultValuesSuite) TestDefaultOutput(c *check.C) {
	output := GetOutput()

	c.Assert(output, check.Equals, os.Stdout)
}

func (s *logTestDefaultValuesSuite) TestDefaultLevel(c *check.C) {
	level := GetLevel()

	c.Assert(level, check.Equals, DebugLevel)
}

type logTestSuite struct {
	output bytes.Buffer
}

var _ = check.Suite(&logTestSuite{})

func (s *logTestSuite) SetUpSuite(c *check.C) {
	SetOutput(&s.output)
}

func (s *logTestSuite) SetUpTest(c *check.C) {
	s.output.Reset()
	SetLevel(DebugLevel)
}

func (s *logTestSuite) TestLogWritesDebugOutput(c *check.C) {
	msg := "this is a debug message"
	Debugf(msg)

	c.Assert(s.output.String(), check.Equals, msg)
}

func (s *logTestSuite) TestLogDoesNotWritesDebugOutputWhenLevelIsInfo(c *check.C) {
	msg := "this is a debug message"

	SetLevel(InfoLevel)
	Debugf(msg)

	c.Assert(s.output.String(), check.Equals, "")
}

func (s *logTestSuite) TestLogWritesDebugOutputWithFormat(c *check.C) {
	msg := "this is a debug message with %d %s"
	par1 := 2
	par2 := "parameters"
	expected := "this is a debug message with 2 parameters"

	Debugf(msg, par1, par2)

	c.Assert(s.output.String(), check.Equals, expected)
}

func (s *logTestSuite) TestLogWritesInfoOutput(c *check.C) {
	msg := "this is a info message"
	Infof(msg)

	c.Assert(s.output.String(), check.Equals, msg)
}

func (s *logTestSuite) TestLogWritesInfoOutputWithFormat(c *check.C) {
	msg := "this is a info message with %d %s"
	par1 := 2
	par2 := "parameters"
	expected := "this is a info message with 2 parameters"

	Infof(msg, par1, par2)

	c.Assert(s.output.String(), check.Equals, expected)
}

func (s *logTestSuite) TestSetTextLevel(c *check.C) {
	currentLvl := GetLevel()
	defer SetLevel(currentLvl)
	testCases := []struct {
		textLvl string
		level   Level
	}{{"info", InfoLevel},
		{"debug", DebugLevel},
	}
	for _, testCase := range testCases {
		err := SetTextLevel(testCase.textLvl)
		lvl := GetLevel()

		c.Check(err, check.IsNil)
		c.Check(lvl, check.Equals, testCase.level)
	}
}

func (s *logTestSuite) TestSetWrongTextLevel(c *check.C) {
	currentLvl := GetLevel()
	defer SetLevel(currentLvl)

	wrongLvl := "not-supported-level"

	err := SetTextLevel(wrongLvl)
	lvl := GetLevel()

	c.Check(err.Error(), check.Equals,
		fmt.Sprintf("The level %s is not supported", wrongLvl))

	c.Check(lvl, check.Equals, currentLvl)
}
