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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/configstate/config"
)

// XXX: not adding anything to "supportedConfigurations" here because
//      this code will only work in the context of firstboot

func handleNetplanConfiguration(tr config.ConfGetter, opts *fsOnlyContext) error {
	// XXX: once we have the full netplan config support do not exit here
	// of course
	if opts == nil {
		return nil
	}

	netplanConfig := filepath.Join(opts.RootDir, "etc/netplan/99-snapd.conf")
	// XXX: should never exist as we run only on fresh images
	if osutil.FileExists(netplanConfig) {
		return fmt.Errorf("cannot write netplan config, %v already exists", netplanConfig)
	}

	// XXX: hack, anything unexpected will be ignored
	// We need something proper for config.ConfGetter like
	// maybe `GetAll("netplan")` or `Unflatten("netplan") or something.
	pcc, ok := tr.(plainCoreConfig)
	if !ok {
		return nil
	}
	// extract net plan config
	netplanCfg := map[string]interface{}{}
	for k, v := range pcc {
		if strings.HasPrefix(k, "netplan.") {
			netplanCfg[k] = v
		}
	}
	if len(netplanCfg) == 0 {
		return nil
	}

	// and write it to the netplan config file
	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	if err := enc.Encode(netplanCfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(netplanConfig), 0755); err != nil {
		return err
	}
	if err := osutil.AtomicWrite(netplanConfig, buf, 0600, 0); err != nil {
		return fmt.Errorf("cannot write network configuration: %v", err)
	}

	return nil
}

func validateNetplanConfiguration(tr config.ConfGetter) error {
	// XXX: do some validation?
	return nil
}
