// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016 Canonical Ltd
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

package builtin

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/snapcore/snapd/interfaces"
)

type ContentSharingInterface struct{}

func (iface *ContentSharingInterface) Name() string {
	return "content"
}

func cleanSubPath(path string) bool {
	return filepath.Clean(path) == path && path != ".." && !strings.HasPrefix(path, "../")
}

func (iface *ContentSharingInterface) SanitizeSlot(slot *interfaces.Slot) error {
	if iface.Name() != slot.Interface {
		panic(fmt.Sprintf("slot is not of interface %q", iface))
	}

	// FIXME: check for read or for write
	rpath, ok := slot.Attrs["read"].([]interface{})
	if !ok || len(rpath) == 0 {
		return fmt.Errorf("content must contain the read attribute")
	}
	for _, r := range rpath {
		if !cleanSubPath(r.(string)) {
			return fmt.Errorf("relative path not allowed")
		}
	}

	return nil
}

func (iface *ContentSharingInterface) SanitizePlug(slot *interfaces.Plug) error {
	if iface.Name() != slot.Interface {
		panic(fmt.Sprintf("plug is not of interface %q", iface))
	}
	// FIXME: hm, check stuff

	return nil
}

func (iface *ContentSharingInterface) ConnectedSlotSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor, interfaces.SecuritySecComp, interfaces.SecurityDBus, interfaces.SecurityUDev, interfaces.SecurityBind:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *ContentSharingInterface) PermanentSlotSnippet(slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {

	switch securitySystem {
	case interfaces.SecurityAppArmor, interfaces.SecuritySecComp, interfaces.SecurityDBus, interfaces.SecurityUDev, interfaces.SecurityBind:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *ContentSharingInterface) ConnectedPlugSnippet(plug *interfaces.Plug, slot *interfaces.Slot, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	contentSnippet := bytes.NewBuffer(nil)
	dst := plug.Attrs["target"].(string)
	dst = filepath.Join(plug.Snap.MountDir(), dst)

	// read
	if readPaths, ok := slot.Attrs["read"].([]interface{}); ok {
		for _, r := range readPaths {
			src := filepath.Join(slot.Snap.MountDir(), r.(string))
			fmt.Fprintf(contentSnippet, "%s %s (ro)\n", src, dst)
		}
	}

	// write
	if writePaths, ok := slot.Attrs["write"].([]interface{}); ok {
		for _, r := range writePaths {
			src := filepath.Join(slot.Snap.MountDir(), r.(string))
			fmt.Fprintf(contentSnippet, "%s %s (rw)\n", src, dst)
		}
	}

	switch securitySystem {
	case interfaces.SecurityBind:
		return contentSnippet.Bytes(), nil
	case interfaces.SecurityAppArmor, interfaces.SecuritySecComp, interfaces.SecurityDBus, interfaces.SecurityUDev:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *ContentSharingInterface) PermanentPlugSnippet(plug *interfaces.Plug, securitySystem interfaces.SecuritySystem) ([]byte, error) {
	switch securitySystem {
	case interfaces.SecurityAppArmor, interfaces.SecuritySecComp, interfaces.SecurityDBus, interfaces.SecurityUDev, interfaces.SecurityBind:
		return nil, nil
	default:
		return nil, interfaces.ErrUnknownSecurity
	}
}

func (iface *ContentSharingInterface) AutoConnect() bool {
	// FIXME: really?
	return true
}
