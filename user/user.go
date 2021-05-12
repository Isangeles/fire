/*
 * user.go
 *
 * Copyright (C) 2020-2021 Dariusz Sikora <dev@isangeles.pl>
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
	"github.com/isangeles/flame/character"
	"github.com/isangeles/flame/flag"

	"github.com/isangeles/fire/data/res"
)

// Struct for user.
type User struct {
	Logged    bool
	admin     bool
	id        string
	pass      string
	charFlags []flag.Flag
	chars     []Character
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
	u.admin = data.Admin
	for _, f := range data.CharFlags {
		u.charFlags = append(u.charFlags, flag.Flag(f))
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

// Admin checks if user is the admin.
func (u *User) Admin() bool {
	return u.admin
}

// Chars returns user characters.
func (u *User) Chars() []Character {
	return u.chars
}

// AddChar adds user's flags to specified character and adds
// this character to the user characters list.
func (u *User) AddChar(c *character.Character) {
	for _, f := range u.charFlags {
		c.AddFlag(f)
	}
	u.chars = append(u.chars, Character{c.ID(), c.Serial()})
}

// CharFlags returns a list of flags that identifies
// the game character as a user character.
func (u *User) CharFlags() []flag.Flag {
	return u.charFlags
}

// Controls checks if user controls object with
// specified ID and serial value.
func (u *User) Controls(id, serial string) bool {
	for _, c := range u.Chars() {
		if c.ID+c.Serial == id+serial {
			return true
		}
	}
	return false
}
