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

package app

import (
	"github.com/ubuntu-core/snappy/systemd"
	"github.com/ubuntu-core/snappy/timeout"
)

// Port is used to declare the Port and Negotiable status of such port
// that is bound to a ServiceYaml.
type Port struct {
	Port       string `yaml:"port,omitempty"`
	Negotiable bool   `yaml:"negotiable,omitempty"`
}

// Ports is a representation of Internal and External ports mapped with a Port.
type Ports struct {
	Internal map[string]Port `yaml:"internal,omitempty" json:"internal,omitempty"`
	External map[string]Port `yaml:"external,omitempty" json:"external,omitempty"`
}

// Yaml represents an application (binary or service)
type Yaml struct {
	// name is partent key
	Name string
	// part of the yaml
	Version string `yaml:"version"`
	Command string `yaml:"command"`
	Daemon  string `yaml:"daemon"`

	Description string          `yaml:"description,omitempty" json:"description,omitempty"`
	Stop        string          `yaml:"stop,omitempty"`
	PostStop    string          `yaml:"poststop,omitempty"`
	StopTimeout timeout.Timeout `yaml:"stop-timeout,omitempty"`
	BusName     string          `yaml:"bus-name,omitempty"`
	Forking     bool            `yaml:"forking,omitempty"`

	// set to yes if we need to create a systemd socket for this service
	Socket       bool   `yaml:"socket,omitempty" json:"socket,omitempty"`
	ListenStream string `yaml:"listen-stream,omitempty" json:"listen-stream,omitempty"`
	SocketMode   string `yaml:"socket-mode,omitempty" json:"socket-mode,omitempty"`
	SocketUser   string `yaml:"socket-user,omitempty" json:"socket-user,omitempty"`
	SocketGroup  string `yaml:"socket-group,omitempty" json:"socket-group,omitempty"`

	// systemd "restart" thing
	RestartCond systemd.RestartCondition `yaml:"restart-condition,omitempty" json:"restart-condition,omitempty"`

	// must be a pointer so that it can be "nil" and omitempty works
	Ports *Ports `yaml:"ports,omitempty" json:"ports,omitempty"`

	OffersRef []string `yaml:"offers"`
	UsesRef   []string `yaml:"uses"`
}
