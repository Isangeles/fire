/*
 * game.go
 *
 * Copyright (C) 2020-2024 Dariusz Sikora <ds@isangeles.dev>
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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/isangeles/flame"
	"github.com/isangeles/flame/area"
	"github.com/isangeles/flame/character"
	flamedata "github.com/isangeles/flame/data"
	flameres "github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/flag"

	"github.com/isangeles/burn/ash"

	"github.com/isangeles/fire/config"
	"github.com/isangeles/fire/data"
	"github.com/isangeles/fire/response"
	"github.com/isangeles/fire/user"
)

const inactiveCharFlag = flag.Flag("flagFireInactive")

// Server-side wrapper for game.
type Game struct {
	*flame.Module
	scripts map[string]*ash.Script
	pause   bool
}

// newGame creates game for specified module data.
// Created game is automatically updated with the frequency
// specified in the config package.
func newGame(data flameres.ModuleData) *Game {
	g := Game{
		Module:  flame.NewModule(data),
		scripts: make(map[string]*ash.Script),
	}
	g.AddChangeChapterEvent(g.changeChapter)
	go g.update()
	err := g.runChapterScripts()
	if err != nil {
		log.Printf("Game: unable to run chapter scripts: %v",
			err)
	}
	return &g
}

// SpawnChar spawns specified character in game start area.
func (g *Game) SpawnChar(char *character.Character) error {
	area := g.Chapter().Area(g.Chapter().Conf().StartArea)
	if area == nil {
		return fmt.Errorf("start area not found: %s",
			g.Chapter().Conf().StartArea)
	}
	area.AddObject(char)
	char.SetPosition(g.Chapter().Conf().StartPosX, g.Chapter().Conf().StartPosY)
	char.SetDestPoint(g.Chapter().Conf().StartPosX, g.Chapter().Conf().StartPosY)
	return nil
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

// UpdateUserChars adds game characters to the specified user according to the
// user configuration and removes characters that don't exists anymore.
func (g *Game) UpdateUserChars(usr *user.User) {
	if len(usr.CharFlags()) < 1 {
		return
	}
	// Add new characters.
outer:
	for _, c := range g.Chapter().Characters() {
		if usr.Controls(c.ID(), c.Serial()) {
			continue
		}
		for _, f := range usr.CharFlags() {
			if !c.HasFlag(f) {
				continue outer
			}
		}
		c.RemoveFlag(inactiveCharFlag)
		usr.AddChar(c)
	}
	// Remove not existing characters.
	for _, char := range usr.Chars() {
		if g.Chapter().Character(char.ID, char.Serial) == nil {
			usr.RemoveChar(char)
		}
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

// StopScripts stops all currently running scripts.
func (g *Game) StopScripts() {
	for _, s := range g.scripts {
		s.Stop(true)
	}
}

// ActivateUserChars removes deactivated char flag from
// all characters of the specified user.
func (g *Game) ActivateUserChars(usr *user.User) {
	for _, c := range g.Chapter().Characters() {
		if usr.Controls(c.ID(), c.Serial()) {
			c.RemoveFlag(inactiveCharFlag)
		}
	}
}

// DeactivatesUserChars add deactivated char flag to all
// characters of the specified user.
func (g *Game) DeactivateUserChars(usr *user.User) {
	for _, c := range g.Chapter().Characters() {
		if usr.Controls(c.ID(), c.Serial()) {
			c.AddFlag(inactiveCharFlag)
		}
	}
}

// NotifyNearChars sends response to all objects that can
// see(have it in sight range) specified area object.
func (g *Game) NotifyNearObjects(ob area.Object, resp response.Response) {
	area := g.Chapter().ObjectArea(ob)
	if area == nil {
		return
	}
	obX, obY := ob.Position()
	for _, ob := range area.SightRangeObjects(obX, obY) {
		charResp := charResponse{
			Response:   resp,
			CharID:     ob.ID(),
			CharSerial: ob.Serial(),
		}
		charResponses <- charResp
	}

}

// ClientData returns game data for server clients.
// Data like characers of incative(offline) users
// will be excluded.
func (g *Game) ClientData() flameres.ModuleData {
	data := g.Data()
	// Search for inactive characters.
	inactiveChars := make(map[string]flameres.CharacterData)
	for _, char := range data.Resources.Characters {
		for _, flag := range char.Flags {
			if flag.ID == inactiveCharFlag.ID() {
				inactiveChars[char.ID+char.Serial] = char
				break
			}
		}
	}
	for _, char := range data.Chapter.Resources.Characters {
		for _, flag := range char.Flags {
			if flag.ID == inactiveCharFlag.ID() {
				inactiveChars[char.ID+char.Serial] = char
				break
			}
		}
	}
	// Exclude invactive characters.
	for id, area := range data.Chapter.Resources.Areas {
		var chars []flameres.AreaCharData
		for _, char := range area.Characters {
			_, inactive := inactiveChars[char.ID+char.Serial]
			if !inactive {
				chars = append(chars, char)
			}
		}
		data.Chapter.Resources.Areas[id].Characters = chars
	}
	return data
}

// update handles game update loop.
func (g *Game) update() {
	update := time.Now()
	for {
		if g.pause {
			continue
		}
		// Delta.
		delta := time.Since(update).Milliseconds()
		// Update.
		g.Module.Update(delta)
		update = time.Now()
		time.Sleep(time.Duration(config.UpdateBreak) * time.Millisecond)
	}
}

// changeChapter handles chapter change triggered by specified character.
func (g *Game) changeChapter(char *character.Character) {
	// Change chapter.
	g.Conf().Chapter = char.ChapterID()
	chapterPath := filepath.Join(g.Conf().ChaptersPath(), g.Conf().Chapter)
	chapterData, err := flamedata.ImportChapterDir(chapterPath)
	if err != nil {
		log.Printf("Unable to change chapter: unable to load chapter data: %v",
			err)
		return
	}
	chapter := flame.NewChapter(g.Module, chapterData)
	g.SetChapter(chapter)
	// Respawn character.
	err = g.SpawnChar(char)
	if err != nil {
		log.Printf("Unable to change chapter: unable to respawn character: %v",
			err)
	}
	// Notify client about chapter change.
	resp := charResponse{
		Response:   response.Response{ChangeChapter: true},
		CharID:     char.ID(),
		CharSerial: char.Serial(),
	}
	charResponses <- resp
}

// runChapterScripts starts all ash scripts for
// current chapter.
func (g *Game) runChapterScripts() error {
	path := filepath.Join(g.Conf().Path, config.ModuleServerPath, "chapters",
		g.Chapter().Conf().ID, "scripts")
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	scripts, err := data.ImportScripts(path)
	if err != nil {
		return fmt.Errorf("unable to import scripts: %v", err)
	}
	for _, s := range scripts {
		go g.runScript(s)
	}
	return nil
}

// runScript runs specified ash script.
func (g *Game) runScript(script *ash.Script) {
	g.scripts[script.Name()] = script
	err := ash.Run(script)
	if err != nil {
		log.Printf("Game: unable to run ash script: %v", err)
	}
	delete(g.scripts, script.Name())
}
