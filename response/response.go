/*
 * response.go
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

// response package provides structs
// for server responses.
package response

import (
	"encoding/json"
	"fmt"

	"github.com/isangeles/flame/data/res"
)

// Struct for server response.
type Response struct {
	Logon          bool                   `json:"logon"`
	Update         Update                 `json:"update"`
	NewChar        []res.CharacterData    `json:"new-char"`
	Trade          []Trade                `json:"trade"`
	TradeCompleted []TradeCompleted       `json:"trade-completed"`
	Dialog         []res.ObjectDialogData `json:"dialog"`
	Command        []Command              `json:"command"`
	Error          []string               `json:"error"`
}

// Unmarshal parses specified text data to response struct.
func Unmarshal(s string) (Response, error) {
	r := Response{}
	err := json.Unmarshal([]byte(s), &r)
	if err != nil {
		return r, fmt.Errorf("unable to unmarshal response: %v",
			err)
	}
	return r, nil
}

// Marshal parses specified response to text data.
func Marshal(r Response) (string, error) {
	out, err := json.Marshal(&r)
	if err != nil {
		return "", fmt.Errorf("unable to marshal response: %v",
			err)
	}
	return string(out[:]), nil
}
