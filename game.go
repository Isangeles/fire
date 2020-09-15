/*
 * game.go
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

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/isangeles/flame"
	flamedata "github.com/isangeles/flame/data"
	flameres "github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/module"
	"github.com/isangeles/flame/module/character"

	"github.com/isangeles/fire/client"
	"github.com/isangeles/fire/config"
)

// Server-side wrapper for game.
type Game struct {
	*flame.Game
}

// newGame loads module with ID set in config and
// starts a game.
func newGame() (*Game, error) {
	if len(config.Module) < 1 {
		return nil, fmt.Errorf("no Flame module configurated")
	}
	modData, err := flamedata.ImportModule(config.ModulePath())
	if err != nil {
		return nil, fmt.Errorf("unable to load game module: %v",
			err)
	}
	mod := module.New(modData)
	g := new(Game)
	g.Game = flame.NewGame(mod)
	return g, nil
}

// Update handles game update loop.
func (g *Game) Update() {
	update := time.Now()
	for {
		// Delta.
		dtNano := time.Since(update).Nanoseconds()
		delta := dtNano / int64(time.Millisecond) // delta to milliseconds
		// Update.
		g.Game.Update(delta)
		update = time.Now()
		time.Sleep(1 * time.Second)
	}
}

// SpawnChar creates new character for specified data and
// spawns it in game start area.
func (g *Game) SpawnChar(data flameres.CharacterData) (*character.Character, error) {
	char := character.New(data)
	chapter := g.Module().Chapter()
	area := chapter.Area(chapter.Conf().StartArea)
	if area == nil {
		return nil, fmt.Errorf("start area not found: %s",
			chapter.Conf().StartArea)
	}
	area.AddCharacter(char)
	char.SetPosition(chapter.Conf().StartPosX, chapter.Conf().StartPosY)
	return char, nil
}

// AddClientChars adds client characters to game.
func (g *Game) AddClientChars(client *client.Client) {
	for _, c := range client.User().Chars {
		charData := flameres.Character(c.ID, c.Serial)
		if charData == nil {
			log.Printf("Client: %s: character not found: %s %s",
				client.RemoteAddr(), c.ID, c.Serial)
			return
		}
		char := character.New(*charData)
		area := game.Module().Chapter().Area(char.AreaID())
		if area == nil {
			log.Printf("Client: %s: character: %s: area not found: %s",
				client.RemoteAddr(), char.ID(), char.AreaID())
			return
		}
		area.AddCharacter(char)
	}
}

// ValidNewCharacter checks if specified data is valid  for the
// new character in current chapter.
func (g *Game) ValidNewCharacter(data flameres.CharacterData) bool {
	if data.Level > g.Module().Chapter().Conf().StartLevel {
		return false
	}
	attrs := data.Attributes.Str + data.Attributes.Con + data.Attributes.Dex +
		data.Attributes.Int + data.Attributes.Wis
	if attrs > g.Module().Chapter().Conf().StartAttrs {
		return false
	}
	for _, i := range data.Inventory.Items {
		for _, id := range g.Module().Chapter().Conf().StartItems {
			if i.ID != id {
				return false
			}
		}
	}
	for _, s := range data.Skills {
		for _, id := range g.Module().Chapter().Conf().StartSkills {
			if s.ID != id {
				return false
			}
		}
	}
	return true
}
