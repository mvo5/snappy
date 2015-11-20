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
	"testing"

	"github.com/ubuntu-core/snappy/dirs"

	. "gopkg.in/check.v1"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type BootloaderTestSuite struct {
}

var _ = Suite(&BootloaderTestSuite{})

type mockBootloader struct {
	BootVars map[string]string
}

func newMockBootloader() *mockBootloader {
	return &mockBootloader{
		BootVars: make(map[string]string),
	}
}
func (b *mockBootloader) Name() bootloaderName {
	return "mocky"
}
func (b *mockBootloader) GetBootVar(name string) (string, error) {
	return b.BootVars[name], nil
}
func (b *mockBootloader) SetBootVar(name, value string) error {
	b.BootVars[name] = value
	return nil
}
func (b *mockBootloader) BootDir() string {
	return ""
}

func (s *BootloaderTestSuite) SetUpTest(c *C) {
	dirs.SetRootDir(c.MkDir())
}

func (s *BootloaderTestSuite) TestMarkBootSuccessfulAllSnap(c *C) {
	b := newMockBootloader()
	bootloader = func() (bootLoader, error) {
		return b, nil
	}

	b.BootVars["snappy_os"] = "os1"
	b.BootVars["snappy_kernel"] = "k1"
	err := MarkBootSuccessful()
	c.Assert(err, IsNil)
	c.Assert(b.BootVars, DeepEquals, map[string]string{
		"snappy_mode":        "regular",
		"snappy_trial_boot":  "0",
		"snappy_kernel":      "k1",
		"snappy_good_kernel": "k1",
		"snappy_os":          "os1",
		"snappy_good_os":     "os1",
	})
}
