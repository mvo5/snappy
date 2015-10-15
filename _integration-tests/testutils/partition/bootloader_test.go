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

package partition

import (
	"fmt"
	"testing"

	"gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

type bootloaderTestSuite struct {
	filepathGlobCalls        map[string]int
	backFilepathGlob         func(string) ([]string, error)
	filepathGlobFail         bool
	filepathGlobReturnValues []string
	fakeConf map[string]string
}

var _ = check.Suite(&bootloaderTestSuite{})

func (s *bootloaderTestSuite) SetUpSuite(c *check.C) {
	s.backFilepathGlob = filepathGlob
	filepathGlob = s.fakeFilepathGlob
	confValue = s.fakeConfValue
	s.fakeConf = map[string]string{}
}

func (s *bootloaderTestSuite) TearDownSuite(c *check.C) {
	filepathGlob = s.backFilepathGlob
	confValue = confValueReal
}

func (s *bootloaderTestSuite) SetUpTest(c *check.C) {
	s.filepathGlobCalls = make(map[string]int)
	s.filepathGlobFail = false
	s.filepathGlobReturnValues = nil
}

func (s *bootloaderTestSuite) fakeFilepathGlob(path string) (matches []string, err error) {
	if s.filepathGlobFail {
		err = fmt.Errorf("Error calling filepathGlob!!")
		return
	}
	s.filepathGlobCalls[path]++

	return s.filepathGlobReturnValues, nil
}

func (s *bootloaderTestSuite) fakeConfValue(key string) (value string, err error) {
	return s.fakeConf[key], nil
}

func (s *bootloaderTestSuite) TestOtherPartition(c *check.C) {
	c.Assert(OtherPartition("a"), check.Equals, "b",
		check.Commentf("Expected OtherPartition of 'a' to be 'b'"))
	c.Assert(OtherPartition("b"), check.Equals, "a",
		check.Commentf("Expected OtherPartition of 'b' to be 'a'"))
}

func (s *bootloaderTestSuite) TestBootDir(c *check.C) {
	c.Assert(BootDir("grub"), check.Equals, grubDir,
		check.Commentf("Expected BootDir of 'grub' to be "+grubDir))
	c.Assert(BootDir("uboot"), check.Equals, ubootDir,
		check.Commentf("Expected OtherPartition of 'uboot' to be "+ubootDir))
}

func (s *bootloaderTestSuite) TestBootSystemReturnsGlobError(c *check.C) {
	s.filepathGlobFail = true

	_, err := BootSystem()

	c.Assert(err, check.NotNil, check.Commentf("Expected error to be nil, %v", err))
}

func (s *bootloaderTestSuite) TestBootSystemCallsFilepathGlob(c *check.C) {
	BootSystem()

	path := bootBase + "/grub"
	calls := s.filepathGlobCalls[path]

	c.Assert(calls, check.Equals, 1,
		check.Commentf("Expected calls to filepath.Glob with path %s to be 1, %d found", path, calls))
}

func (s *bootloaderTestSuite) TestBootSystemForGrub(c *check.C) {
	s.filepathGlobReturnValues = []string{"a-grub-related-dir"}

	bootSystem, err := BootSystem()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(bootSystem, check.Equals, "grub",
		check.Commentf("Expected grub boot system not found, %s", bootSystem))
}

func (s *bootloaderTestSuite) TestBootSystemForUBoot(c *check.C) {
	s.filepathGlobReturnValues = []string{}

	bootSystem, err := BootSystem()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(bootSystem, check.Equals, "uboot",
		check.Commentf("Expected uboot boot system not found, %s", bootSystem))
}

func (s *bootloaderTestSuite) TestNextBootPartitionReturnsBootSystemError(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "try",
		"snappy_ab": "dummy",
	}
	
	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	expectedErr := fmt.Errorf("Error from BootSystem!")
	BootSystem = func() (system string, err error) {
		return "", expectedErr
	}

	_, err := NextBootPartition()

	c.Assert(err, check.Equals, expectedErr,
		check.Commentf("Expected error %v not found, %v", expectedErr, err))
}

func (s *bootloaderTestSuite) TestNextBootPartitionReturnsEmptyIfPatternsNotFound(c *check.C) {
	s.fakeConf = map[string]string{"snappy_mode": "try"}

	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	BootSystem = func() (system string, err error) {
		return "test-system", nil
	}

	partition, err := NextBootPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(partition, check.Equals, "",
		check.Commentf("NextBootPartition %s, expected empty", partition))
}

func (s *bootloaderTestSuite) TestNextBootPartitionReturnsSamePartitionForGrub(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "try",
		"snappy_ab": "a",
	}

	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	BootSystem = func() (system string, err error) {
		return "grub", nil
	}

	partition, err := NextBootPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(partition, check.Equals, "a",
		check.Commentf("NextBootPartition %s, expected a", partition))
}

func (s *bootloaderTestSuite) TestNextBootPartitionReturnsOtherPartitionForUBoot(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "try",
		"snappy_ab": "a",
	}

	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	BootSystem = func() (system string, err error) {
		return "uboot", nil
	}

	partition, err := NextBootPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(partition, check.Equals, "b",
		check.Commentf("NextBootPartition %s, expected b", partition))
}

func (s *bootloaderTestSuite) TestModeReturnsSnappyModeFromConf(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "test_mode",
	}

	mode, err := Mode()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(mode, check.Equals, "test_mode", check.Commentf("Wrong mode"))
}

func (s *bootloaderTestSuite) TestCurrentPartitionNotOnTryMode(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "nottry",
		"snappy_ab": "test_partition",
	}
	
	part, err := CurrentPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(part, check.Equals, "test_partition", check.Commentf("Wrong partition"))
}

func (s *bootloaderTestSuite) TestCurrentPartitionOnTryModeReturnsSamePartitionForUboot(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "try",
		"snappy_ab": "a",
	}

	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	BootSystem = func() (system string, err error) {
		return "uboot", nil
	}

	mode, err := CurrentPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(mode, check.Equals, "a", check.Commentf("Wrong partition"))
}

func (s *bootloaderTestSuite) TestCurrentPartitionOnTryModeReturnsOtherPartitionForGrub(c *check.C) {
	s.fakeConf = map[string]string{
		"snappy_mode": "try",
		"snappy_ab": "a",
	}

	backBootSystem := BootSystem
	defer func() { BootSystem = backBootSystem }()
	BootSystem = func() (system string, err error) {
		return "grub", nil
	}

	mode, err := CurrentPartition()

	c.Assert(err, check.IsNil, check.Commentf("Unexpected error %v", err))
	c.Assert(mode, check.Equals, "b", check.Commentf("Wrong partition"))
}
