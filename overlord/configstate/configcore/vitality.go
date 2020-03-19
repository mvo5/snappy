// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2020 Canonical Ltd
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

package configcore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/configstate/config"
	"github.com/snapcore/snapd/progress"
	"github.com/snapcore/snapd/snap/naming"
	"github.com/snapcore/snapd/strutil"
	"github.com/snapcore/snapd/systemd"
)

const vitalityOpt = "resiliance.vitality"
const baseScore = 899

func init() {
	// add supported configuration of this module
	supportedConfigurations["core."+vitalityOpt] = true
}

func generateVitalityOverrideContent(rank int) []byte {
	s := fmt.Sprintf(`[Service]
OOMScoreAdjust=-%d
`, baseScore-rank)
	return []byte(s)
}

var (
	// XXX: make better
	snapNameMatcher    = regexp.MustCompile(`.*/snap.(.*).*.service.d/vitality.conf`)
	snapServiceMatcher = regexp.MustCompile(`.*/(snap.*\.service)(.d/vitality.conf)?`)
)

func handleVitalityConfiguration(tr config.Conf) error {
	option, err := coreCfg(tr, vitalityOpt)
	if err != nil {
		return err
	}
	if option == "" {
		return nil
	}

	var needsRestart []string

	// remove old overrides
	snaps := strings.Split(option, ",")
	vitalityScoreFiles, err := filepath.Glob(filepath.Join(dirs.SnapServicesDir, "snap.*.service.d/vitality.conf"))
	if err != nil {
		return nil
	}
	for _, p := range vitalityScoreFiles {
		m := snapNameMatcher.FindStringSubmatch(p)
		if len(m) < 2 {
			return fmt.Errorf("internal error: cannot find match for %s", p)
		}
		if !strutil.ListContains(snaps, m[1]) {
			if err := os.Remove(p); err != nil {
				return err
			}
			// ensure we restart to adjust oom
			m := snapServiceMatcher.FindStringSubmatch(p)
			if len(m) < 2 {
				return fmt.Errorf("internal error: cannot find service name from %s", p)
			}
			needsRestart = append(needsRestart, m[1])
		}
	}

	// add new overrides
	for i, name := range strings.Split(option, ",") {
		// XXX: fugly, can we do something else?
		snapSystemdServices, err := filepath.Glob(filepath.Join(dirs.SnapServicesDir, fmt.Sprintf("snap.%s.*.service", name)))
		if err != nil {
			return err
		}

		for _, p := range snapSystemdServices {
			appOverrideDir := p + ".d"
			if err := os.MkdirAll(appOverrideDir, 0755); err != nil {
				return err
			}
			overridePath := filepath.Join(appOverrideDir, "vitality.conf")
			// XXX: can we use osutil.EnsureDirState() here?
			err = osutil.EnsureFileState(overridePath, &osutil.MemoryFileState{
				Content: generateVitalityOverrideContent(i),
				Mode:    0644,
			})
			if err == osutil.ErrSameState {
				continue
			}
			if err != nil {
				return err
			}
			// ensure we restart to adjust oom
			m := snapServiceMatcher.FindStringSubmatch(p)
			if len(m) < 2 {
				return fmt.Errorf("internal error: cannot find service name from %s", p)
			}
			needsRestart = append(needsRestart, m[1])
		}
	}

	if len(needsRestart) > 0 {
		sysd := systemd.New(dirs.GlobalRootDir, systemd.SystemMode, progress.Null)
		if err := sysd.DaemonReload(); err != nil {
			return err
		}
		for _, srv := range needsRestart {
			// best effort
			if b, _ := sysd.IsActive(srv); b {
				// XXX: what is the right interval?
				sysd.Restart(srv, 30*time.Second)
			}
		}
	}

	return nil
}

func validateVitalitySettings(tr config.Conf) error {
	option, err := coreCfg(tr, vitalityOpt)
	if err != nil {
		return err
	}
	if option == "" {
		return nil
	}

	for _, snapName := range strings.Split(option, ",") {
		if err := naming.ValidateInstance(snapName); err != nil {
			return fmt.Errorf("cannot set %q: %v", vitalityOpt, err)
		}
	}

	return nil
}
