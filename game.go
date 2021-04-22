/*
 * game.go
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

package main

import (
	"fmt"
	"time"

	"github.com/isangeles/flame"
	"github.com/isangeles/flame/character"
	flameres "github.com/isangeles/flame/data/res"

	"github.com/isangeles/fire/config"
	"github.com/isangeles/fire/user"
)

// Server-side wrapper for game.
type Game struct {
	*flame.Module
}

// newGame creates game for specified module data.
func newGame(data flameres.ModuleData) *Game {
	mod := flame.NewModule()
	mod.Apply(data)
	g := Game{Module: mod}
	return &g
}

// Update handles game update loop.
func (g *Game) Update() {
	update := time.Now()
	for {
		// Delta.
		dtNano := time.Since(update).Nanoseconds()
		delta := dtNano / int64(time.Millisecond) // delta to milliseconds
		// Update.
		g.Module.Update(delta)
		update = time.Now()
		time.Sleep(time.Duration(config.UpdateBreak) * time.Millisecond)
	}
}

// SpawnChar creates new character for specified data and
// spawns it in game start area.
func (g *Game) SpawnChar(data flameres.CharacterData) (*character.Character, error) {
	char := character.New(data)
	chapter := g.Chapter()
	area := chapter.Area(chapter.Conf().StartArea)
	if area == nil {
		return nil, fmt.Errorf("start area not found: %s",
			chapter.Conf().StartArea)
	}
	area.AddCharacter(char)
	char.SetPosition(chapter.Conf().StartPosX, chapter.Conf().StartPosY)
	return char, nil
}

// ValidNewCharacter checks if specified data is valid  for the
// new character in current chapter.
func (g *Game) ValidNewCharacter(data flameres.CharacterData) bool {
	if data.Level > g.Chapter().Conf().StartLevel {
		return false
	}
	attrs := data.Attributes.Str + data.Attributes.Con + data.Attributes.Dex +
		data.Attributes.Int + data.Attributes.Wis
	if attrs > g.Chapter().Conf().StartAttrs {
		return false
	}
	for _, i := range data.Inventory.Items {
		for _, id := range g.Chapter().Conf().StartItems {
			if i.ID != id {
				return false
			}
		}
	}
	for _, s := range data.Skills {
		for _, id := range g.Chapter().Conf().StartSkills {
			if s.ID != id {
				return false
			}
		}
	}
	return true
}

// AddUserChars add game characters to the specified user according to the
// user configuration.
func (g *Game) AddUserChars(usr *user.User) {
	if len(usr.CharFlags()) < 1 {
		return
	}
outer:
	for _, c := range g.Chapter().Characters() {
		if usr.Controls(c.ID(), c.Serial()) {
			continue
		}
		for _, f := range usr.CharFlags() {
			if !c.HasFlag(f) {
				break outer
			}
		}
		usr.Chars = append(usr.Chars, user.Character{c.ID(), c.Serial()})
	}
}

// AddTranslationAll adds specified translation to all
// existing translation bases in the game module.
func (g *Game) AddTranslationAll(data flameres.TranslationData) {
	res := g.Resources()
	for i, _ := range res.TranslationBases {
		res.TranslationBases[i].Translations = append(res.TranslationBases[i].Translations, data)
	}
}
