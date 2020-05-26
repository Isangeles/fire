/*
 * confirm.go
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

	"github.com/isangeles/flame/module/character"

	"github.com/isangeles/fire/client"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/response"
)

// handleConfirmedRequest handles specified request as confirmed.
func handleConfirmedRequest(req charConfirmRequest) {
	resp := response.Response{}
	for _, t := range req.Trade {
		tradeAccept := response.TradeAccepted{
			ID:           req.ID,
			BuyerID:      t.BuyerID,
			BuyerSerial:  t.BuyerSerial,
			SellerID:     t.SellerID,
			SellerSerial: t.SellerSerial,
			ItemsBuy:     t.ItemsBuy,
			ItemsSell:    t.ItemsSell,
		}
		resp.TradeAccepted = append(resp.TradeAccepted, tradeAccept)
		handleConfirmedTradeRequest(req.Client, t, &resp)
	}
	req.Client.Out <- resp
}

// handleConfirmedTradeRequest handles specified trade request as confirmed.
func handleConfirmedTradeRequest(cli *client.Client, req request.Trade, resp *response.Response) {
	// Find buyer and seller.
	object := game.Module().Object(req.BuyerID, req.BuyerSerial)
	if object == nil {
		err := fmt.Sprintf("Object not found: %s %s", req.BuyerID, req.BuyerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	buyer, ok := object.(*character.Character)
	if !ok {
		err := fmt.Sprintf("Object is not a character: %s %s", req.BuyerID,
			req.BuyerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	object = game.Module().Object(req.SellerID, req.SellerSerial)
	if object == nil {
		err := fmt.Sprintf("Object not found: %s %s", req.SellerID,
			req.SellerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	seller, ok := object.(*character.Character)
	if !ok {
		err := fmt.Sprintf("Object is not a character: %s %s", req.SellerID,
			req.SellerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Trade items.
	err := transferItems(seller, buyer, req.ItemsBuy)
	if err != nil {
		err := fmt.Sprintf("Unable to transfer items to buy: %v", err)
		resp.Errors = append(resp.Errors, err)
		return
	}
	err = transferItems(buyer, seller, req.ItemsSell)
	if err != nil {
		err := fmt.Sprintf("Unable to transfer items to sell: %v", err)
		resp.Errors = append(resp.Errors, err)
		return
	}
}
