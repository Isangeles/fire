/*
 * request.go
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

// request package provides structs
// for client requests.
package request

import (
	"encoding/json"
	"fmt"

	"github.com/isangeles/flame/data/res"
)

// Struct for client request.
type Request struct {
	Login         []Login             `json:"login"`
	NewChar       []res.CharacterData `json:"new-char"`
	Move          []Move              `json:"move"`
	Dialog        []Dialog            `json:"dialog"`
	DialogAnswer  []DialogAnswer      `json:"dialog-answer"`
	Trade         []Trade             `json:"trade"`
	TransferItems []TransferItems     `json:"transfer-items"`
	Skill         []Skill             `json:"skill"`
	Use           []Use               `json:"use"`
	Accept        []int               `json:"accept"`
	Command       []string            `json:"command"`
}

// Unmarshal parses specified text data to action struct.
func Unmarshal(data string) (*Request, error) {
	a := new(Request)
	err := json.Unmarshal([]byte(data), a)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal request: %v",
			err)
	}
	return a, nil
}

// Marshal parses specified action to text data.
func Marshal(req *Request) (string, error) {
	out, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("unable to marshal request: %v", err)
	}
	return string(out[:]), nil
}
