// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
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

package selftest_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"

	"github.com/snapcore/snapd/selftest"
)

// Hook up check.v1 into the "go test" runner
func Test(t *testing.T) { TestingT(t) }

type selftestSuite struct{}

var _ = Suite(&selftestSuite{})

func (s *selftestSuite) TestRunHappy(c *C) {
	var happyChecks []func() error
	var happyCheckRan int

	happyChecks = append(happyChecks, func() error {
		happyCheckRan += 1
		return nil
	})

	restore := selftest.MockChecks(happyChecks)
	defer restore()

	err := selftest.Run()
	c.Check(err, IsNil)
	c.Check(happyCheckRan, Equals, 1)
}

func (s *selftestSuite) TestRunNotHappy(c *C) {
	var unhappyChecks []func() error
	var unhappyCheckRan int

	unhappyChecks = append(unhappyChecks, func() error {
		unhappyCheckRan += 1
		return fmt.Errorf("unhappy")
	})

	restore := selftest.MockChecks(unhappyChecks)
	defer restore()

	p := filepath.Join(c.MkDir(), "/run/snapd/selftest")
	oldSelftestResult := dirs.SelftestResult
	defer func() { dirs.SelftestResult = oldSelftestResult }()
	dirs.SelftestResult = p

	err := selftest.Run()
	c.Check(err, ErrorMatches, "unhappy")
	c.Check(unhappyCheckRan, Equals, 1)

	content, err := ioutil.ReadFile(p)
	c.Assert(err, IsNil)
	c.Check(string(content), Equals, "unhappy")
}
