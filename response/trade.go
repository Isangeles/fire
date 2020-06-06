/*
 * trade.go
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

package response

// Struct for trade response.
type Trade struct {
	ID           int                 `json:"id"`
	BuyerID      string              `json:"buyer-id"`
	BuyerSerial  string              `json:"buyer-serial"`
	SellerID     string              `json:"seller-id"`
	SellerSerial string              `json:"seller-serial"`
	ItemsBuy     map[string][]string `json:"items-buy"`
	ItemsSell    map[string][]string `json:"items-sell"`
}

// Struct for trade completed response.
type TradeCompleted Trade
