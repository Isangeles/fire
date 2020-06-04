/*
 * utils.go
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

	"github.com/isangeles/flame/module/item"
	"github.com/isangeles/flame/module/objects"

	"github.com/isangeles/fire/config"
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

// inRange checks if specified objects are in the minimum range between each other.
// The minimum range value is specified in config package.
// The function always returns true if at least one of the specified objects have
// no position.
func inRange(ob1, ob2 objects.Object) bool {
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
