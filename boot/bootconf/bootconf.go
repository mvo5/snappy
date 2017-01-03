// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2015-2016 Canonical Ltd
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

package bootconf

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/snapcore/snapd/osutil"
)

var piBootConfig = "/boot/uboot/config.txt"

// FIXME: add regexp to validate values here as well?
var piValidKeys = []string{
	"disable_overscan",
	"framebuffer_width",
	"framebuffer_height",
	"framebuffer_depth",
	"framebuffer_ignore_alpha",
	"overscan_left",
	"overscan_right",
	"overscan_top",
	"overscan_bottom",
	"overscan_scale",
	"display_rotate",
	"hdmi_group",
	"hdmi_mode",
	"hdmi_drive",
	"avoid_warnings",
	"gpu_mem_256",
	"gpu_mem_512",
	"gpu_mem",
}

type Pi struct {
	Path string
}

func (pi *Pi) validateKey(key string) error {
	for _, valid := range piValidKeys {
		if valid == key {
			return nil
		}
	}

	return fmt.Errorf("cannot use boot config key %q", key)
}

func (pi *Pi) Get(keys []string) (map[string]string, error) {
	for _, key := range keys {
		if err := pi.validateKey(key); err != nil {
			return nil, err
		}
	}

	f, err := os.Open(pi.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		for _, key := range keys {
			line := scanner.Text()
			if strings.HasPrefix(line, fmt.Sprintf("%s=", key)) {
				l := strings.SplitN(line, "=", 2)
				out[l[0]] = l[1]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (pi *Pi) Set(nv map[string]string) error {
	for key, _ := range nv {
		if err := pi.validateKey(key); err != nil {
			return err
		}
	}

	f, err := os.Open(pi.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		for key, val := range nv {
			if strings.HasPrefix(line, fmt.Sprintf("%s=", key)) || strings.HasPrefix(line, fmt.Sprintf("#%s=", key)) {
				line = fmt.Sprintf("%s=%s", key, val)
				delete(nv, key)
				break
			}
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// add leftovers
	for key, val := range nv {
		lines = append(lines, fmt.Sprintf("%s=%s", key, val))
	}

	return osutil.AtomicWriteFile(pi.Path, []byte(strings.Join(lines, "\n")), 0644, 0)
}
