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

package devicestate

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"gopkg.in/tomb.v2"

	"github.com/snapcore/snapd/boot"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/snapstate"
	"github.com/snapcore/snapd/overlord/state"
	"github.com/snapcore/snapd/timings"
)

func (m *DeviceManager) doSetupRunSystem(t *state.Task, _ *tomb.Tomb) error {
	st := t.State()
	st.Lock()
	defer st.Unlock()

	perfTimings := timings.NewForTask(t)
	defer perfTimings.Save(st)

	// get gadget dir
	deviceCtx, err := DeviceCtx(st, t, nil)
	if err != nil {
		return fmt.Errorf("cannot get device context: %v", err)
	}
	info, err := snapstate.GadgetInfo(st, deviceCtx)
	if err != nil {
		return fmt.Errorf("cannot get gadget info: %v", err)
	}
	gadgetDir := info.MountDir()

	// run the create partition code
	st.Unlock()
	output, err := exec.Command(filepath.Join(dirs.DistroLibExecDir, "snap-bootstrap"), "create-partitions", gadgetDir).CombinedOutput()
	st.Lock()
	if err != nil {
		return osutil.OutputErr(output, err)
	}

	// now make the "run" system bootable
	bootWith := &boot.BootableSet{
		UnpackedGadgetDir: gadgetDir,

		// XXX: better way to override this
		UC20InstallMode: true,
	}
	model := deviceCtx.Model()
	if err := boot.MakeBootable(model, dirs.GlobalRootDir, bootWith); err != nil {
		return err
	}
	// XXX: set mode to "run" in recovery grub.cfg

	st.RequestRestart(state.RestartSystem)
	return nil
}
