/*
 * equip.go
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

package request

// Struct for equip request.
type Equip struct {
	CharID     string          `json:"char-id"`
	CharSerial string          `json:"char-serial"`
	ItemID     string          `json:"item-id"`
	ItemSerial string          `json:"item-serial"`
	Slots      []EquipmentSlot `json:"slot"`
}

// Struct for unequip request.
type Unequip struct {
	CharID     string `json:"char-id"`
	CharSerial string `json:"char-serial"`
	ItemID     string `json:"item-id"`
	ItemSerial string `json:"item-serial"`
}

// Struct for equipment slot data.
type EquipmentSlot struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}
