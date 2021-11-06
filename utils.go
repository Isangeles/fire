/*
 * utils.go
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
	"strings"

	"github.com/isangeles/flame/character"
	flamedata "github.com/isangeles/flame/data"
	flameres "github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/item"
	"github.com/isangeles/flame/objects"
	"github.com/isangeles/flame/serial"
	"github.com/isangeles/flame/useaction"

	"github.com/isangeles/fire/config"
	"github.com/isangeles/fire/request"
)

// transferItems transfer items between specified objects.
// Items are in the form of a map with IDs as keys and serial values as values.
func transferItems(from, to item.Container, items map[string][]string) error {
	for id, serials := range items {
		for _, serial := range serials {
			item := from.Inventory().Item(id, serial)
			if item == nil {
				return fmt.Errorf("Item not found: %s %s",
					id, serial)
			}
			from.Inventory().RemoveItem(item)
			err := to.Inventory().AddItem(item)
			if err != nil {
				return fmt.Errorf("Unable to add item inventory: %v",
					err)
			}
		}
	}
	return nil
}

// charSkillRecipe returns skill or recipe with specified ID from the character,
// or nil if character does not have skill or recipe with such ID.
func charSkillRecipe(char *character.Character, id string) useaction.Usable {
	for _, s := range char.Skills() {
		if s.ID() == id {
			return s
		}
	}
	for _, r := range char.Crafting().Recipes() {
		if r.ID() == id {
			return r
		}
	}
	return nil
}

// inRange checks if specified objects are in the minimum range between each other.
// The minimum range value is specified in config package.
// The function always returns true if at least one of the specified objects have
// no position.
func inRange(ob1, ob2 serial.Serialer) bool {
	pos1, ok := ob1.(objects.Positioner)
	if !ok {
		return true
	}
	pos2, ok := ob2.(objects.Positioner)
	if !ok {
		return true
	}
	return objects.Range(pos1, pos2) <= config.ActionMinRange
}

// equip inserts item to specified slots in character equipment.
func equip(eq *character.Equipment, it item.Equiper, slots []request.EquipmentSlot) error {
	for _, slot := range slots {
		for _, eqSlot := range eq.Slots() {
			if string(eqSlot.Type()) != slot.Type || eqSlot.ID() != slot.ID {
				continue
			}
			for _, itSlot := range it.Slots() {
				if itSlot == eqSlot.Type() {
					eqSlot.SetItem(it)
				}
			}
		}
	}
	for _, itSlot := range it.Slots() {
		equiped := false
		for _, eqSlot := range eq.Slots() {
			if eqSlot.Type() != itSlot || eqSlot.Item() != it {
				continue
			}
			equiped = true
			break
		}
		if !equiped {
			eq.Unequip(it)
			return fmt.Errorf("Item is was not inserted in all required slots: %s %s: %s",
				it.ID(), it.Serial(), string(itSlot))
		}
	}
	return nil
}

// importModule imports module data from module directory or module file.
func importModule(path string) (flameres.ModuleData, error) {
	if strings.HasSuffix(path, flamedata.ModuleFileExt) {
		return flamedata.ImportModule(path)
	}
	return flamedata.ImportModuleDir(path)
}
