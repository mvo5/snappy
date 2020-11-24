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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SnapCtlOptions holds the various options with which snapctl is invoked.
type SnapCtlOptions struct {
	// ContextID is a string used to determine the context of this call (e.g.
	// which context and handler should be used, etc.)
	ContextID string `json:"context-id"`

	// Args contains a list of parameters to use for this invocation.
	Args []string `json:"args"`
}

type snapctlOutput struct {
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
}

// RunSnapctl requests a snapctl run for the given options.
func (client *Client) RunSnapctl(stdin io.ReadCloser, options *SnapCtlOptions) (stdout, stderr []byte, err error) {
	b, err := json.Marshal(options)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot marshal options: %s", err)
	}

	stdinErrCh := make(chan error)
	go func() {
		// we cannot use golang http here, the issue is that golang
		// tries to read the POST "body" there is no EOF from stdin
		// so this needs to implemented manually
		hcl := client.doer.(*http.Client)
		tr := hcl.Transport.(*http.Transport)
		conn, err := tr.Dial("", "")
		if err != nil {
			stdinErrCh <- err
		}

		// XXX: content-lengt is needed here or go will not read body
		d := fmt.Sprintf("POST /v2/snapctl/stdin/%s HTTP/1.1\nHost: localhost\nContent-Length: 999999\n\n", options.ContextID)
		fmt.Println(d)
		n, err := conn.Write([]byte(d))
		fmt.Println("written", n, err)
		if err != nil {
			stdinErrCh <- err
		}
		go func() {
			io.Copy(conn, stdin)
		}()
		var buf [512]byte
		n, err = conn.Read(buf[:])
		fmt.Println("got", n, err, string(buf[:]))

		stdinErrCh <- err
	}()

	var output snapctlOutput
	_, err = client.doSync("POST", "/v2/snapctl", nil, nil, bytes.NewReader(b), &output)
	if err != nil {
		return nil, nil, err
	}
	err = <-stdinErrCh

	return []byte(output.Stdout), []byte(output.Stderr), err
}
