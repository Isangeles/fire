/*
 * request.go
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

// request package provides structs
// for client requests.
package request

import (
	"encoding/json"
	"fmt"
)

// Struct for client request.
type Request struct {
	Login         []Login         `json:"login"`
	NewChar       []NewChar       `json:"new-char"`
	SetPos        []SetPos        `json:"set-pos"`
	Move          []Move          `json:"move"`
	Dialog        []Dialog        `json:"dialog"`
	DialogAnswer  []DialogAnswer  `json:"dialog-answer"`
	DialogEnd     []DialogEnd     `json:"dialog-end"`
	Trade         []Trade         `json:"trade"`
	TransferItems []TransferItems `json:"transfer-items"`
	ThrowItems    []ThrowItems    `json:"throw-items"`
	Use           []Use           `json:"use"`
	Equip         []Equip         `json:"equip"`
	Unequip       []Unequip       `json:"unequip"`
	Training      []Training      `json:"training"`
	Chat          []Chat          `json:"chat"`
	Target        []Target        `json:"target"`
	Save          []string        `json:"save"`
	Load          string          `json:"load"`
	Command       []string        `json:"command"`
	Accept        []int           `json:"accept"`
	Close         int64           `json:"close"`
	Pause         bool            `json:"pause"`
}

// Unmarshal parses specified text data to action struct.
func Unmarshal(data string) (*Request, error) {
	req := new(Request)
	err := json.Unmarshal([]byte(data), req)
	if err != nil {
		return nil, fmt.Errorf("unable to unmarshal request: %v",
			err)
	}
	return req, nil
}

// Marshal parses specified action to text data.
func Marshal(req *Request) (string, error) {
	out, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("unable to marshal request: %v", err)
	}
	return string(out[:]), nil
}
