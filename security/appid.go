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

package security

import (
	"bytes"
	"fmt"
	"strings"
)

type AppID struct {
	AppID   string
	Pkgname string
	Appname string
	Version string
}

func NewAppID(appID string) (*AppID, error) {
	tmp := strings.Split(appID, "_")
	if len(tmp) != 3 {
		return nil, ErrInvalidAppID
	}
	id := AppID{
		AppID:   appID,
		Pkgname: tmp[0],
		Appname: tmp[1],
		Version: tmp[2],
	}
	return &id, nil
}

// TODO: once verified, reorganize all these
func (sa *AppID) AppArmorVars() string {
	aavars := fmt.Sprintf(`
# Specified profile variables
@{APP_APPNAME}="%s"
@{APP_ID_DBUS}="%s"
@{APP_PKGNAME_DBUS}="%s"
@{APP_PKGNAME}="%s"
@{APP_VERSION}="%s"
@{INSTALL_DIR}="{/snaps,/gadget}"
# Deprecated:
@{CLICK_DIR}="{/snaps,/gadget}"`, sa.Appname, dbusPath(sa.AppID), dbusPath(sa.Pkgname), sa.Pkgname, sa.Version)
	return aavars
}

const allowed = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789`

// Generate a string suitable for use in a DBus object
func dbusPath(s string) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(s)))

	for _, c := range []byte(s) {
		if strings.IndexByte(allowed, c) >= 0 {
			fmt.Fprintf(buf, "%c", c)
		} else {
			fmt.Fprintf(buf, "_%02x", c)
		}
	}

	return buf.String()
}
