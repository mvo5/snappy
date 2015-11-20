package snappy

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

import (
	"regexp"

	"github.com/ubuntu-core/snappy/partition"
)

// used in the unit tests
var getBootVar = partition.GetBootVar

var bootVarSpliter = regexp.MustCompile(`(.*)_(.*).snap`)

func rebootRequiredPartsForBootvar(nextBootName, goodBootName string) ([]Part, error) {
	nextBoot, err := getBootVar(nextBootName)
	if err != nil {
		return nil, err
	}
	goodBoot, err := getBootVar(goodBootName)
	if err != nil {
		return nil, err
	}

	if nextBoot == goodBoot {
		return nil, nil
	}

	resultSlice := bootVarSpliter.FindStringSubmatch(nextBoot)
	name := resultSlice[1]
	ver := resultSlice[2]
	installed, err := NewMetaRepository().Installed()
	if err != nil {
		return nil, err
	}

	return FindSnapsByNameAndVersion(name, ver, installed), nil
}

// requireRebootParts returns the parts that require a reboot in
// oder to become active. These are the OS and kernel snaps
func requireRebootParts() (result []Part, err error) {
	// we have the kernel or os that may require a reboot
	k, err := rebootRequiredPartsForBootvar("snappy_kernel", "snappy_good_kernel")
	if err != nil {
		return nil, err
	}
	result = append(result, k...)
	os, err := rebootRequiredPartsForBootvar("snappy_os", "snappy_good_os")
	if err != nil {
		return nil, err
	}
	result = append(result, os...)

	return result, nil
}
