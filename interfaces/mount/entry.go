// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package mount

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall"
)

// Entry describes an /etc/fstab-like mount entry.
//
// Fields are named after names in struct returned by getmntent(3).
//
// struct mntent {
//     char *mnt_fsname;   /* name of mounted filesystem */
//     char *mnt_dir;      /* filesystem path prefix */
//     char *mnt_type;     /* mount type (see Mntent.h) */
//     char *mnt_opts;     /* mount options (see Mntent.h) */
//     int   mnt_freq;     /* dump frequency in days */
//     int   mnt_passno;   /* pass number on parallel fsck */
// };
type Entry struct {
	Name    string
	Dir     string
	Type    string
	Options []string

	DumpFrequency   int
	CheckPassNumber int
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EqualEntries checks if one entry is equal to another
func (a *Entry) Equal(b *Entry) bool {
	return (a.Name == b.Name && a.Dir == b.Dir && a.Type == b.Type &&
		equalStrings(a.Options, b.Options) && a.DumpFrequency == b.DumpFrequency &&
		a.CheckPassNumber == b.CheckPassNumber)
}

// escape replaces whitespace characters so that getmntent can parse it correctly.
var escape = strings.NewReplacer(
	" ", `\040`,
	"\t", `\011`,
	"\n", `\012`,
	"\\", `\134`,
).Replace

// unescape replaces escape sequences used by setmnt with whitespace characters.
var unescape = strings.NewReplacer(
	`\040`, " ",
	`\011`, "\t",
	`\012`, "\n",
	`\134`, "\\",
).Replace

func (e Entry) String() string {
	// Name represents name of the device in a mount entry.
	name := "none"
	if e.Name != "" {
		name = escape(e.Name)
	}
	// Dir represents mount directory in a mount entry.
	dir := "none"
	if e.Dir != "" {
		dir = escape(e.Dir)
	}
	// Type represents file system type in a mount entry.
	fsType := "none"
	if e.Type != "" {
		fsType = escape(e.Type)
	}
	// Options represents mount options in a mount entry.
	options := "defaults"
	if len(e.Options) != 0 {
		options = escape(strings.Join(e.Options, ","))
	}
	return fmt.Sprintf("%s %s %s %s %d %d",
		name, dir, fsType, options, e.DumpFrequency, e.CheckPassNumber)
}

// ParseEntry parses a fstab-like entry.
func ParseEntry(s string) (Entry, error) {
	var e Entry
	var err error
	var df, cpn int
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '\t' })
	// do all error checks before any assignments to `e'
	if len(fields) < 4 || len(fields) > 6 {
		return e, fmt.Errorf("expected between 4 and 6 fields, found %d", len(fields))
	}
	// Parse DumpFrequency if we have at least 5 fields
	if len(fields) > 4 {
		df, err = strconv.Atoi(fields[4])
		if err != nil {
			return e, fmt.Errorf("cannot parse dump frequency: %q", fields[4])
		}
	}
	// Parse CheckPassNumber if we have at least 6 fields
	if len(fields) > 5 {
		cpn, err = strconv.Atoi(fields[5])
		if err != nil {
			return e, fmt.Errorf("cannot parse check pass number: %q", fields[5])
		}
	}
	e.Name = unescape(fields[0])
	e.Dir = unescape(fields[1])
	e.Type = unescape(fields[2])
	e.Options = strings.Split(unescape(fields[3]), ",")
	e.DumpFrequency = df
	e.CheckPassNumber = cpn
	return e, nil
}

// LoadFSTab reads and parses an fstab-like file.
//
// The supported format is described by fstab(5).
func LoadFSTab(reader io.Reader) ([]Entry, error) {
	var entries []Entry
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s := scanner.Text()
		if i := strings.IndexByte(s, '#'); i != -1 {
			s = s[0:i]
		}
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		entry, err := ParseEntry(s)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

// SaveFSTab writes a list of entries to a fstab-like file.
//
// The supported format is described by fstab(5).
//
// Note that there is no support for comments, both the LoadFSTab function and
// SaveFSTab just ignore them.
//
// Note that there is no attempt to use atomic file write/rename tricks. The
// created file will typically live in /run/snapd/ns/$SNAP_NAME.fstab and will
// be done so, while holidng a flock-based-lock, by the snap-update-ns program.
func SaveFSTab(writer io.Writer, entries []Entry) error {
	var buf bytes.Buffer
	for i := range entries {
		if _, err := fmt.Fprintf(&buf, "%s\n", entries[i]); err != nil {
			return err
		}
	}
	_, err := buf.WriteTo(writer)
	return err
}

// OptsToFlags converts mount options strings to a mount flag.
func OptsToFlags(opts []string) (flags int, err error) {
	for _, opt := range opts {
		switch opt {
		case "ro":
			flags |= syscall.MS_RDONLY
		case "nosuid":
			flags |= syscall.MS_NOSUID
		case "nodev":
			flags |= syscall.MS_NODEV
		case "noexec":
			flags |= syscall.MS_NOEXEC
		case "sync":
			flags |= syscall.MS_SYNCHRONOUS
		case "remount":
			flags |= syscall.MS_REMOUNT
		case "mand":
			flags |= syscall.MS_MANDLOCK
		case "dirsync":
			flags |= syscall.MS_DIRSYNC
		case "noatime":
			flags |= syscall.MS_NOATIME
		case "nodiratime":
			flags |= syscall.MS_NODIRATIME
		case "bind":
			flags |= syscall.MS_BIND
		case "rbind":
			flags |= syscall.MS_BIND | syscall.MS_REC
		case "move":
			flags |= syscall.MS_MOVE
		case "silent":
			flags |= syscall.MS_SILENT
		case "acl":
			flags |= syscall.MS_POSIXACL
		case "private":
			flags |= syscall.MS_PRIVATE
		case "rprivate":
			flags |= syscall.MS_PRIVATE | syscall.MS_REC
		case "slave":
			flags |= syscall.MS_SLAVE
		case "rslave":
			flags |= syscall.MS_SLAVE | syscall.MS_REC
		case "shared":
			flags |= syscall.MS_SHARED
		case "rshared":
			flags |= syscall.MS_SHARED | syscall.MS_REC
		case "relatime":
			flags |= syscall.MS_RELATIME
		case "strictatime":
			flags |= syscall.MS_STRICTATIME
		default:
			return 0, fmt.Errorf("unsupported mount option: %q", opt)
		}
	}
	return flags, nil
}
