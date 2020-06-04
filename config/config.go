/*
 * config.go
 *
 * Copyright (C) 2020 Dariusz Sikora <dev@isangeles.pl>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

// Package for handling server
// configuration file.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/isangeles/flame/data/text"
)

const (
	Name, Version  = "Fire", "0.0.0"
	ConfigFileName = ".fire"
)

var (
	Host           = ""
	Port           = "8000"
	Module         = ""
	ActionMinRange = 50.0
)

// LoadConfig load server configuration file.
func Load() error {
	// Open config file.
	file, err := os.Open(ConfigFileName)
	if err != nil {
		return fmt.Errorf("unable to open config file: %v", err)
	}
	defer file.Close()
	// Unmarshal config.
	conf, err := text.UnmarshalConfig(file)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config: %v", err)
	}
	if len(conf["host"]) > 0 {
		Host = conf["host"][0]
	}
	if len(conf["port"]) > 0 {
		Port = conf["port"][0]
	}
	if len(conf["module"]) > 0 {
		Module = conf["module"][0]
	}
	if len(conf["action-min-range"]) > 0 {
		minRange, err := strconv.ParseFloat(conf["action-min-range"][0], 64)
		if err == nil {
			ActionMinRange = minRange
		}
	}
	return nil
}

// SaveConfig saves server configuration file.
func Save() error {
	// Create config file.
	file, err := os.Create(ConfigFileName)
	if err != nil {
		return fmt.Errorf("unable to create file: %v", err)
	}
	defer file.Close()
	conf := make(map[string][]string)
	conf["host"] = []string{Host}
	conf["port"] = []string{Port}
	conf["module"] = []string{Module}
	conf["action-min-range"] = []string{fmt.Sprintf("%f", ActionMinRange)}
	text := text.MarshalConfig(conf)
	// Write config to file.
	write := bufio.NewWriter(file)
	write.WriteString(text)
	write.Flush()
	return nil
}

// ModulePath returns path to current module directory.
func ModulePath() string {
	return filepath.Join("data/modules", Module)
}

// UsersDir returns path to directory with users.
func UsersDir() string {
	return filepath.FromSlash("data/users")
}
