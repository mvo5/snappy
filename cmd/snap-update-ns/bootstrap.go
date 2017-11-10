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

package main

// Use a pre-main helper to switch the mount namespace. This is required as
// golang creates threads at will and setns(..., CLONE_NEWNS) fails if any
// threads apart from the main thread exist.

/*

#include <stdlib.h>
#include "bootstrap.h"

__attribute__((section(".preinit_array"), used)) static typeof(&bootstrap) init = &bootstrap;

// NOTE: do not add anything before the following `import "C"'
*/
import "C"

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

var (
	// ErrNoNamespace is returned when a snap namespace does not exist.
	ErrNoNamespace = errors.New("cannot update mount namespace that was not created yet")
)

// IMPORTANT: all the code in this section may be run with elevated privileges
// when invoking snap-update-ns from the setuid snap-confine.

// BootstrapError returns error (if any) encountered in pre-main C code.
func BootstrapError() error {
	if C.bootstrap_msg == nil {
		return nil
	}
	errno := syscall.Errno(C.bootstrap_errno)
	// Translate EINVAL from setns or ENOENT from open into a dedicated error.
	if errno == syscall.EINVAL || errno == syscall.ENOENT {
		return ErrNoNamespace
	}
	if errno != 0 {
		return fmt.Errorf("%s: %s", C.GoString(C.bootstrap_msg), errno)
	}
	return fmt.Errorf("%s", C.GoString(C.bootstrap_msg))
}

// END IMPORTANT

func makeArgv(args []string) []*C.char {
	argv := make([]*C.char, len(args)+1)
	for i, arg := range args {
		argv[i] = C.CString(arg)
	}
	return argv
}

func freeArgv(argv []*C.char) {
	for _, arg := range argv {
		C.free(unsafe.Pointer(arg))
	}
}

// findSnapName parses the argv-like array and finds the 1st argument.
func findSnapName(args []string) *string {
	argv := makeArgv(args)
	defer freeArgv(argv)

	if ptr := C.find_snap_name(C.int(len(args)), &argv[0]); ptr != nil {
		str := C.GoString(ptr)
		return &str
	}
	return nil
}

// findFirstOption returns the first "-option" string in argv-like array.
func hasOption(args []string, opt string) bool {
	argv := makeArgv(args)
	defer freeArgv(argv)
	cOpt := C.CString(opt)
	defer C.free(unsafe.Pointer(cOpt))

	found := C.has_option(C.int(len(args)), &argv[0], cOpt)
	return bool(found)
}

// validateSnapName checks if snap name is valid.
// This also sets bootstrap_msg on failure.
func validateSnapName(snapName string) int {
	cStr := C.CString(snapName)
	defer C.free(unsafe.Pointer(cStr))
	return int(C.validate_snap_name(cStr))
}

// processArguments parses commnad line arguments.
// The argument cmdline is a string with embedded
// NUL bytes, separating particular arguments.
func processArguments(args []string) (snapName string, shouldSetNs bool) {
	argv := makeArgv(args)
	defer freeArgv(argv)

	var snapNameOut *C.char
	var shouldSetNsOut C.bool
	C.process_arguments(C.int(len(args)), &argv[0], &snapNameOut, &shouldSetNsOut)
	if snapNameOut != nil {
		snapName = C.GoString(snapNameOut)
	}
	shouldSetNs = bool(shouldSetNsOut)

	return snapName, shouldSetNs
}
