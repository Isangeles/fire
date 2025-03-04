/*
 * fire.go
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

// main package handles incoming connections
// and server clients.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	flameres "github.com/isangeles/flame/data/res"
	flamelog "github.com/isangeles/flame/log"
	"github.com/isangeles/flame/serial"

	"github.com/isangeles/burn"

	"github.com/isangeles/fire/config"
	"github.com/isangeles/fire/data"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/response"
)

var (
	game            *Game
	enter           = make(chan *Client)
	leave           = make(chan string)
	requests        = make(chan clientRequest)
	charResponses   = make(chan charResponse)
	confirmRequests = make(chan charConfirmRequest)
	confirmed       = make(chan *clientConfirm)
	load            = make(chan response.Load)
	pendingReqs     = make(map[int]charConfirmRequest)
	close           bool
)

// Server-side wrapper for client request.
type clientRequest struct {
	*request.Request
	Client *Client
}

// Struct with request for character to
// confirm.
type charConfirmRequest struct {
	clientRequest
	CharID     string
	CharSerial string
	ID         int
}

// Struct for client confirmation.
type clientConfirm struct {
	ID     int
	Client *Client
}

// Struct with response for the owner of game
// character with specifed ID and serial.
type charResponse struct {
	response.Response
	CharID     string
	CharSerial string
}

// Main function.
func main() {
	flamelog.PrintStdOut = true
	err := config.Load()
	if err != nil {
		err := config.Save()
		if err != nil {
			log.Printf("Unable to save defualt config: %v", err)
		}
	}
	err = data.LoadUsers(config.UsersPath)
	if err != nil {
		log.Printf("Unable to load users: %v", err)
	}
	if len(config.Module) < 1 {
		panic(fmt.Errorf("No game module configurated"))
	}
	modData, err := importModule(config.ModulePath())
	if err != nil {
		panic(fmt.Errorf("Unable to load game module: %v", err))
	}
	game = newGame(modData)
	burn.Module = game.Module
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		panic(fmt.Errorf("Unable to create listener: %v", err))
	}
	log.Printf("%s(%s)@%s", config.Name, config.Version, server.Addr())
	go update()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %v", err)
		}
		go handleConnection(conn)
	}
}

// update handles client enter/leave, requests and
// communication between clients.
func update() {
	clients := make(map[string]*Client)
	for {
		select {
		case user := <-enter:
			clients[user.RemoteAddr().String()] = user
			user.Out <- response.Response{Logon: true}
			log.Printf("Enters: %s", user.RemoteAddr())
		case addr := <-leave:
			client := clients[addr]
			if client == nil {
				continue
			}
			if client.User() != nil {
				game.DeactivateUserChars(client.User())
			}
			client.Close()
			delete(clients, addr)
			log.Printf("Leaves: %s", addr)
		case req := <-requests:
			handleRequest(req)
		case resp := <-charResponses:
			for _, c := range clients {
				if c.User() != nil && !c.User().Controls(resp.CharID, resp.CharSerial) {
					continue
				}
				updateClient(c, resp.Response)
				break
			}
		case req := <-confirmRequests:
			pendingReqs[req.ID] = req
		case con := <-confirmed:
			req := pendingReqs[con.ID]
			if con.Client != nil && !con.Client.User().Controls(req.CharID, req.CharSerial) {
				continue
			}
			handleConfirmedRequest(req)
			delete(pendingReqs, int(con.ID))
		case resp := <-load:
			flameres.Clear()
			serial.Reset()
			game.StopScripts()
			game = newGame(resp.Module)
			for _, c := range clients {
				if c.User() == nil {
					continue
				}
				updateClient(c, response.Response{Load: resp})
			}
		}
		for _, c := range clients {
			if c.User() == nil {
				continue
			}
			updateClient(c, response.Response{})
		}
		if close {
			// Wait some time before closing the server to ensure that
			// all clients will receive closed response.
			time.AfterFunc(time.Duration(2) * time.Second, closeServer)
		}
	}
}

// handleConnection handles client connection.
func handleConnection(conn net.Conn) {
	// Create client.
	cli := newClient(conn)
	defer cli.Close()
	// Start client writer.
	go clientWriter(cli)
	// Enter & listen.
	enter <- cli
	input := bufio.NewScanner(cli)
	for input.Scan() {
		if input.Err() != nil {
			log.Printf("Client: %s: unable to read input: %v",
				cli.RemoteAddr(), input.Err())
		}
		r, err := request.Unmarshal(input.Text())
		if err != nil {
			log.Printf("Client: %s: unable to create request: %v",
				cli.RemoteAddr(), err)
			resp := response.Response{
				Logon: cli.User() == nil,
				Error: []string{"Invalid request syntax"},
			}
			cli.Out <- resp
			continue
		}
		req := clientRequest{r, cli}
		requests <- req
	}
	// Leave.
	leave <- cli.RemoteAddr().String()
}

// updateClient adds update, logon, closed, and charcter responses to specified
// response and sends this response to the specified client.
func updateClient(client *Client, resp response.Response) {
	// Update user characters.
	if client.User() != nil {
		game.UpdateUserChars(client.User())
		for _, c := range client.User().Chars() {
			charResp := response.Character{c.ID, c.Serial}
			resp.Character = append(resp.Character, charResp)
		}
	}
	// Send update response.
	resp.Update = response.Update{Module: game.ClientData(), Message: config.Message}
	resp.Logon = client.User() == nil
	resp.Closed = close
	resp.Paused = game.pause
	client.Out <- resp
}

// clientWriter handles writing on client out channel.
func clientWriter(c *Client) {
	for r := range c.Out {
		respData, err := response.Marshal(r)
		if err != nil {
			log.Printf("Client writer: %s: unable to marshal server response: %v",
				c.RemoteAddr(), err)
			return
		}
		_, err = fmt.Fprintf(c, "%s\r\n", respData)
		if err != nil {
			log.Printf("Client writer: %s: unable to write on client out: %v",
				c.RemoteAddr(), err)
		}
	}
}

// closeServer saves current server configuration and terminates the program.
func closeServer() {
	err := config.Save()
	if err != nil {
		log.Printf("Unable to save config: %v", err)
	}
	os.Exit(0)
}
