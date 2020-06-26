/*
 * user.go
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

// Package with server user struct.
package user

import (
	"github.com/isangeles/fire/data/res"
)

// Struct for user.
type User struct {
	id     string
	pass   string
	Logged bool
	Chars  []Character
}

// Struct for user character.
type Character struct {
	ID, Serial string
}

// New creates new user.
func New(data res.UserData) *User {
	u := new(User)
	u.id = data.ID
	u.pass = data.Pass
	for i, s := range data.Chars {
		u.Chars = append(u.Chars, Character{i, s})
	}
	return u
}

// ID returns user ID.
func (u *User) ID() string {
	return u.id
}

// Pass returns user password.
func (u *User) Pass() string {
	return u.pass
}

// Controls checks if user controls object with
// specified ID and serial value.
func (u *User) Controls(id, serial string) bool {
	for _, c := range u.Chars {
		if c.ID + c.Serial == id + serial {
			return true
		}
	}
	return false
}
