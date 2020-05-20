/*
 * dialog.go
 *
 * Copyright 2020 Dariusz Sikora <dev@isangeles.pl>
 *
 * This program is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston,
 * MA 02110-1301, USA.
 *
 *
 */

package request

// Struct for dialog request.
type Dialog struct {
	TargetID     string `json:"target-id"`
	TargetSerial string `json:"target-serial"`
	OwnerID      string `json:"owner-id"`
	OwnerSerial  string `json:"owner-serial"`
	DialogID     string `json:"dialog-id"`
}

// Struct for dialog answer request.
type DialogAnswer struct {
	Dialog
	AnswerID     string `json:"answer-id"`
}
