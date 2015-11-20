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

// Package partition manipulate snappy disk partitions
package partition

import (
	"errors"
)

const (
	// Name of writable user data partition label as created by
	// ubuntu-device-flash(1).
	writablePartitionLabel = "writable"

	// Name of primary root filesystem partition label as created by
	// ubuntu-device-flash(1).
	rootfsAlabel = "system-a"

	// Name of primary root filesystem partition label as created by
	// ubuntu-device-flash(1). Note that this partition will
	// only be present if this is an A/B upgrade system.
	rootfsBlabel = "system-b"

	// name of boot partition label as created by ubuntu-device-flash(1).
	bootPartitionLabel = "system-boot"

	// File creation mode used when any directories are created
	dirMode = 0750
)

var (
	// ErrBootloader is returned if the bootloader can not be determined
	ErrBootloader = errors.New("Unable to determine bootloader")

	// ErrPartitionDetection is returned if the partition type can not
	// be detected
	ErrPartitionDetection = errors.New("Failed to detect system type")

	// ErrNoDualPartition is returned if you try to use a dual
	// partition feature on a single partition
	ErrNoDualPartition = errors.New("No dual partition")
)
