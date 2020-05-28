/*
 * client.go
 *
 * Copyright (C) 2020 Dariusz Sikora <<dev@isangeles.pl>>
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
 * Copyright 2020 Dariusz Sikora <dev@isangeles.pl>
 *
 */

package client

import (
	"net"

	"github.com/isangeles/fire/response"
	"github.com/isangeles/fire/user"
)

// Struct for client.
type Client struct {
	net.Conn
	user *user.User
	Out  chan response.Response
}

// New makes new client from
// specified connection.
func New(conn net.Conn) *Client {
	c := new(Client)
	c.Conn = conn
	c.Out = make(chan response.Response, 2)
	return c
}

// User returns client user.
func (c *Client) User() *user.User {
	return c.user
}

// SetUser sets client user
func (c *Client) SetUser(u *user.User) {
	u.Logged = true
	c.user = u
}

// Close closes client connection and removes
// a logged flag from client user.
func (c *Client) Close() {
	if c.User() != nil {
		c.User().Logged = false
	}
	c.Conn.Close()
}
