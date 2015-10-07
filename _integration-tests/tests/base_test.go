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

package tests

import (
	"io"
	"log"
	"os"
	"testing"

	"launchpad.net/snappy/_integration-tests/testutils/config"
	"launchpad.net/snappy/_integration-tests/testutils/report"
	"launchpad.net/snappy/_integration-tests/testutils/runner"
	"launchpad.net/snappy/_integration-tests/testutils/tlog"
)

func init() {
	cfg, err := config.ReadConfig(
		"_integration-tests/data/output/testconfig.json")

	if err != nil {
		log.Panic(err.Error())
	}

	if err = tlog.SetTextLevel(cfg.LogLevel); err != nil {
		log.Panic(err.Error())
	}
}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	output := io.MultiWriter(
		os.Stdout,
		&report.ParserReporter{
			Next: &report.FileReporter{}})

	runner.TestingT(t, output)
}
