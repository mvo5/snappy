// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
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
package recover

import (
	"fmt"

	"github.com/snapcore/snapd/cmd/snap-recovery/partition"
	"github.com/snapcore/snapd/gadget"
)

type Options struct {
	EncryptDataPartition bool
}

func Run(gadgetRoot, device string, options *Options) error {
	if gadgetRoot == "" {
		return fmt.Errorf("cannot use empty recovery gadget root directory")
	}
	if device == "" {
		return fmt.Errorf("cannot use empty device node")
	}
	if options == nil {
		options = &Options{}
	}

	// XXX: ensure we test that the current partition table is
	//      compatible with the gadget
	lv, err := gadget.PositionedVolumeFromGadget(gadgetRoot)
	if err != nil {
		return fmt.Errorf("cannot layout the volume: %v", err)
	}

	sfdisk := partition.NewSFDisk(device)
	created, err := sfdisk.Create(lv)
	if err != nil {
		return fmt.Errorf("cannot create the partitions: %v", err)
	}
	// XXX: we need an integration style tests here now to ensure that
	// the filesystem creation is called in the right way
	for _, part := range created {
		if options.EncryptDataPartition && part.Role == "system-data" {
			// system-data roles are always called ubuntu-data for now
			partLabel := "ubuntu-data"
			if err := partition.MakeEncrypted(&part, partLabel); err != nil {
				return err
			}
		}

		if err := partition.MakeFilesystem(part); err != nil {
			return err
		}

		if err := partition.DeployContent(part, gadgetRoot); err != nil {
			return err
		}
	}

	return nil
}
