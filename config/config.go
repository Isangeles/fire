/*
 * config.go
 *
 * Copyright (C) 2020-2025 Dariusz Sikora <ds@isangeles.dev>
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
	Name, Version    = "Fire", "0.1.0-dev"
	ConfigFileName   = ".fire"
	ModulesPath      = "data/modules"
	UsersPath        = "data/users"
	ModuleServerPath = "fire" // path to the server directory inside module directory
)

var (
	Host           = ""
	Port           = "8000"
	Module         = ""
	UpdateBreak    = 1
	ActionMinRange = 50.0
	Message        = ""
)

// Load load server configuration file.
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
	if len(conf["update-break"]) > 0 {
		updateBreak, err := strconv.Atoi(conf["update-break"][0])
		if err == nil {
			UpdateBreak = updateBreak
		}
	}
	if len(conf["action-min-range"]) > 0 {
		minRange, err := strconv.ParseFloat(conf["action-min-range"][0], 64)
		if err == nil {
			ActionMinRange = minRange
		}
	}
	if len(conf["message"]) > 0 {
		Message = conf["message"][0]
	}
	return nil
}

// Save saves server configuration file.
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
	conf["update-break"] = []string{fmt.Sprintf("%d", UpdateBreak)}
	conf["action-min-range"] = []string{fmt.Sprintf("%f", ActionMinRange)}
	conf["message"] = []string{Message}
	text := text.MarshalConfig(conf)
	// Write config to file.
	write := bufio.NewWriter(file)
	write.WriteString(text)
	write.Flush()
	return nil
}

// ModulePath returns path to the current module directory.
func ModulePath() string {
	return filepath.Join(ModulesPath, Module)
}
