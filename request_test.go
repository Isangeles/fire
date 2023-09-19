/*
 * request_test.go
 *
 * Copyright (C) 2022-2023 Dariusz Sikora <<ds@isangeles.dev>>
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
	"testing"

	"github.com/isangeles/flame/character"
	flameres "github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/item"
	"github.com/isangeles/flame/skill"
	"github.com/isangeles/flame/training"

	"github.com/isangeles/fire/data/res"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/user"
)

var (
	skillData = flameres.SkillData{
		ID: "skill",
		UseAction: flameres.UseActionData{
			TargetEffects: []flameres.UseActionEffectData{flameres.UseActionEffectData{}},
		},
	}
	trainingData = flameres.TrainingData{
		ID: "trainingSkill",
		Use: flameres.UseActionData{
			UserMods: flameres.ModifiersData{
				AddSkillMods: []flameres.AddSkillModData{flameres.AddSkillModData{"skill"}},
			},
		},
	}
	trainerTrainingData = flameres.TrainerTrainingData{
		ID: "training",
	}
	itemData = flameres.MiscItemData{ID: "item"}
	charData = flameres.CharacterData{
		ID:         "char",
		Level:      1,
		Attributes: flameres.AttributesData{5, 5, 5, 5, 5},
	}
	areaData      = flameres.AreaData{ID: "area"}
	resourcesData = flameres.ResourcesData{Areas: []flameres.AreaData{areaData}}
	chapterData   = flameres.ChapterData{ID: "chapter", Resources: resourcesData}
	modData       = flameres.ModuleData{ID: "module", Chapter: chapterData}
	userData      = res.UserData{ID: "user"}
)

// TestHandleTransferItemsRequest tests handling transfer items request.
func TestHandleTransferItemsRequest(t *testing.T) {
	// Create game.
	game = newGame(modData)
	// Create characters.
	charFromData := charData
	charFromData.ID = "charFrom"
	charFrom := character.New(charFromData)
	charFrom.SetHealth(0)
	charToData := charData
	charToData.ID = "charTo"
	charTo := character.New(charToData)
	// Add chars to area.
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	area.AddObject(charFrom)
	area.AddObject(charTo)
	// Add items.
	game.Update(1)
	item1 := item.NewMisc(itemData)
	item2 := item.NewMisc(itemData)
	charFrom.Inventory().AddItem(item1)
	charFrom.Inventory().AddItem(item2)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(charTo)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.TransferItems{
		ObjectFromID:     charFrom.ID(),
		ObjectFromSerial: charFrom.Serial(),
		ObjectToID:       charTo.ID(),
		ObjectToSerial:   charTo.Serial(),
		Items:            make(map[string][]string),
	}
	req.Items[item1.ID()] = []string{item1.Serial(), item2.Serial()}
	// Test out of range.
	charTo.SetPosition(100, 100)
	err := handleTransferItemsRequest(client, req)
	if err == nil {
		t.Errorf("Request handling didn't returned out of range error")
	}
	// Test in range.
	charTo.SetPosition(0, 0)
	err = handleTransferItemsRequest(client, req)
	if err != nil {
		t.Fatalf("Request handing error: %v", err)
	}
	if charFrom.Inventory().Item(item1.ID(), item1.Serial()) != nil {
		t.Errorf("Item should be removed from %s inventory: %s %s", charFrom.ID(),
			item1.ID(), item1.Serial())
	}
	if charFrom.Inventory().Item(item2.ID(), item2.Serial()) != nil {
		t.Errorf("Item should be removed from %s inventory: %s %s", charFrom.ID(),
			item2.ID(), item2.Serial())
	}
	if charTo.Inventory().Item(item1.ID(), item1.Serial()) == nil {
		t.Errorf("Item should be added to %s inventory: %s %s", charTo.ID(),
			item1.ID(), item1.Serial())
	}
	if charTo.Inventory().Item(item2.ID(), item2.Serial()) == nil {
		t.Errorf("Item should be added to %s inventory: %s %s", charTo.ID(),
			item2.ID(), item2.Serial())
	}
}

// TestHandleThrowItemsRequest tests handling throw items request.
func TestHandleThrowItemsRequest(t *testing.T) {
	// Create game & character.
	game = newGame(modData)
	char := character.New(charData)
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	area.AddObject(char)
	// Add items.
	game.Update(1)
	item1 := item.NewMisc(itemData)
	item2 := item.NewMisc(itemData)
	char.Inventory().AddItem(item1)
	char.Inventory().AddItem(item2)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(char)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.ThrowItems{
		ObjectID:     char.ID(),
		ObjectSerial: char.Serial(),
		Items:        make(map[string][]string),
	}
	req.Items[item1.ID()] = []string{item1.Serial(), item2.Serial()}
	// Test.
	err := handleThrowItemsRequest(client, req)
	if err != nil {
		t.Fatalf("Request handing error: %v", err)
	}
	if char.Inventory().Item(item1.ID(), item1.Serial()) != nil {
		t.Errorf("Item should be removed from %s inventory: %s %s", char.ID(),
			item1.ID(), item1.Serial())
	}
	if char.Inventory().Item(item2.ID(), item2.Serial()) != nil {
		t.Errorf("Item should be removed from %s inventory: %s %s", char.ID(),
			item2.ID(), item2.Serial())
	}
}

// TestHandleTrainingRequest tests handling training request.
func TestHandleTrainingRequest(t *testing.T) {
	// Create game & character.
	game = newGame(modData)
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	char := character.New(charData)
	area.AddObject(char)
	// Add training skill to resources base.
	flameres.Skills = append(flameres.Skills, skillData)
	// Create trainer.
	trainer := character.New(charData)
	area.AddObject(trainer)
	train := training.New(trainingData)
	trainerTrain := training.NewTrainerTraining(train, trainerTrainingData)
	trainer.AddTraining(trainerTrain)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(char)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.Training{
		UserID:        char.ID(),
		UserSerial:    char.Serial(),
		TrainerID:     trainer.ID(),
		TrainerSerial: trainer.Serial(),
		TrainingID:    trainerTrain.ID(),
	}
	// Test out of range.
	trainer.SetPosition(100, 100)
	err := handleTrainingRequest(client, req)
	if err == nil {
		t.Errorf("Request handling didn't returned out of range error")
	}
	// Test in range.
	trainer.SetPosition(0, 0)
	err = handleTrainingRequest(client, req)
	if err != nil {
		t.Fatalf("Request handing error: %v", err)
	}
	game.Update(1)
	if charSkillRecipe(char, skillData.ID) == nil {
		t.Fatalf("Skill not trained: %s", skillData.ID)
	}
}

// TestHandleUseRequestSkill tests handling use request for skill.
func TestHandleUseRequestSkill(t *testing.T) {
	// Create game & character.
	game = newGame(modData)
	char := character.New(charData)
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	area.AddObject(char)
	// Add skills.
	skill := skill.New(skillData)
	char.AddSkill(skill)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(char)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.Use{
		ObjectID:     skill.ID(),
		ObjectSerial: "",
		UserID:       char.ID(),
		UserSerial:   char.Serial(),
	}
	// Test.
	err := handleUseRequest(client, req)
	if err != nil {
		t.Fatalf("Request handing error: %v", err)
	}
	game.Update(1)
	if char.Cooldown() <= 0 {
		t.Fatalf("Skill was not used: character cooldown: %d", char.Cooldown())
	}
}

// TestHandleUseRequestAreaObject tests handling use request for area object.
func TestHandleUseRequestAreaObject(t *testing.T) {
	// Create game & objects.
	game = newGame(modData)
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	char := character.New(charData)
	area.AddObject(char)
	ob := character.New(charData)
	area.AddObject(ob)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(char)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.Use{
		ObjectID:     ob.ID(),
		ObjectSerial: ob.Serial(),
		UserID:       char.ID(),
		UserSerial:   char.Serial(),
	}
	// Test out of range.
	ob.SetPosition(100, 100)
	err := handleUseRequest(client, req)
	if err == nil {
		t.Errorf("Request handing didn't returned out of range error")
	}
	// Test in range.
	ob.SetPosition(0, 0)
	err = handleUseRequest(client, req)
	if err != nil {
		t.Fatalf("Request handling error: %v", err)
	}
	game.Update(1)
	if char.Cooldown() <= 0 {
		t.Fatalf("Skill was not used: character cooldown: %d", char.Cooldown())
	}
}

// TestHandleChatRequest tests handling chat request.
func TestHandleChatRequest(t *testing.T) {
	// Create game & character.
	game = newGame(modData)
	char := character.New(charData)
	area := game.Chapter().Area("area")
	if area == nil {
		t.Fatalf("Test area not found")
	}
	area.AddObject(char)
	// Create user & client.
	user := user.New(userData)
	user.AddChar(char)
	client := new(Client)
	client.SetUser(user)
	// Create request.
	req := request.Chat{
		ObjectID:     char.ID(),
		ObjectSerial: char.Serial(),
		Message:      "Test message",
	}
	// Test.
	err := handleChatRequest(client, req)
	if err != nil {
		t.Fatalf("Request handing error: %v", err)
	}
	msg := char.ChatLog().Messages()[0]
	if msg.Text != req.Message {
		t.Fatalf("Request message was not added to the chat log")
	}
}
