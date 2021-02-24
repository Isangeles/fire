/*
 * ai.go
 *
 * Copyright (C) 2021 Dariusz Sikora <dev@isangeles.pl>
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

	"github.com/isangeles/fire/response"

	"github.com/isangeles/flame/module/character"
)

// handleAICharResponse handles response send to the
// character controlled by AI.
func handleAICharResponse(resp charResponse) {
	for _, r := range resp.Trade {
		handleAICharTradeResponse(r)
	}
}

// handleAICharTradeResponse handles trade response send
// to the character cntrolled by AI.
func handleAICharTradeResponse(resp response.Trade) error {
	// Find seller & buyer.
	object := game.Module().Object(resp.SellerID, resp.SellerSerial)
	if object == nil {
		return fmt.Errorf("Seller not found: %s %s", resp.SellerID,
			resp.SellerSerial)
	}
	seller, ok := object.(*character.Character)
	if !ok {
		return fmt.Errorf("Seller is not a character: %s %s", resp.SellerID,
			resp.SellerSerial)
	}
	object = game.Module().Object(resp.BuyerID, resp.BuyerSerial)
	if object == nil {
		return fmt.Errorf("Buyer not found: %s %s", resp.BuyerID,
			resp.BuyerSerial)
	}
	buyer, ok := object.(*character.Character)
	if !ok {
		return fmt.Errorf("Buyer is not a character: %s %s", resp.BuyerID,
			resp.BuyerSerial)
	}
	// Validate trade.
	buyValue := 0
	for id, serials := range resp.ItemsBuy {
		for _, serial := range serials {
			it := seller.Inventory().Item(id, serial)
			if it != nil {
				buyValue += it.Value()
			}
		}
	}
	sellValue := 0
	for id, serials := range resp.ItemsSell {
		for _, serial := range serials {
			it := buyer.Inventory().Item(id, serial)
			if it != nil {
				sellValue += it.Value()
			}
		}
	}
	if sellValue < buyValue {
		return nil
	}
	confirm := clientConfirm{resp.ID, nil}
	confirmReq := func() { confirmed <- &confirm }
	go confirmReq()
	return nil
}
