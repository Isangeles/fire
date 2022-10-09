/*
 * chat.go
 *
 * Copyright (C) 2022 Dariusz Sikora <ds@isangeles.dev>
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

package response

import "time"

// Struct for chat request.
type Chat struct {
	ObjectID     string    `json:"object-id"`
	ObjectSerial string    `json:"object-serial"`
	Message      string    `json:"message"`
	Translated   bool      `json:"translated"`
	Time         time.Time `json:"time"`
}
