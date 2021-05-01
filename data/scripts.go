/*
 * scripts.go
 *
 * Copyright (C) 2021 Dariusz Sikora <dev@isangeles.pl>
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

package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"log"

	"github.com/isangeles/burn/ash"
)

const (
	ashScriptExt = ".ash"
)

// ImportScripts imports all scripts form directory with
// specified path.
func ImportScripts(path string) ([]*ash.Script, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read dir: %v", err)
	}
	scripts := make([]*ash.Script, 0)
	for _, info := range files {
		if !strings.HasSuffix(info.Name(), ashScriptExt) {
			continue
		}
		scriptPath := filepath.Join(path, info.Name())
		script, err := ImportScript(scriptPath)
		if err != nil {
			log.Printf("Data: unable to retrieve script: %v",
				err)
		}
		scripts = append(scripts, script)
	}
	return scripts, nil
}

// ImportScript imports script from file with specified path.
func ImportScript(path string) (*ash.Script, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	text, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}
	scriptName := filepath.Base(path)
	script, err := ash.NewScript(scriptName, fmt.Sprintf("%s", text))
	if err != nil {
		return nil, fmt.Errorf("unable to create script: %v", err)
	}
	return script, nil
}
