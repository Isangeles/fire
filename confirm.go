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
		err := handleConfirmedTradeRequest(req.Client, t)
		if err != nil {
			err := fmt.Sprintf("Unable to handle trade request: %v", err)
			resp.Errors = append(resp.Errors, err)
			continue
		}
		tradeCompleted := response.TradeCompleted{
			ID:           req.ID,
			BuyerID:      t.Buy.ObjectToID,
			BuyerSerial:  t.Buy.ObjectToSerial,
			SellerID:     t.Sell.ObjectToID,
			SellerSerial: t.Sell.ObjectToSerial,
			ItemsBuy:     t.Buy.Items,
			ItemsSell:    t.Sell.Items,
		}
		resp.TradeCompleted = append(resp.TradeCompleted, tradeCompleted)
		// Send trade completed response to seller.
		charResp := charResponse{
			CharID:     t.Sell.ObjectToID,
			CharSerial: t.Sell.ObjectToSerial,
		}
		charResp.Response.TradeCompleted = append(charResp.Response.TradeCompleted,
			tradeCompleted)
		sendCharResp := func() { charResponses <- charResp }
		go sendCharResp()
	}
	req.Client.Out <- resp
}

// handleConfirmedTradeRequest handles specified trade request as confirmed.
func handleConfirmedTradeRequest(cli *client.Client, req request.Trade) error {
	// Find buyer.
	object := game.Module().Object(req.Buy.ObjectToID, req.Buy.ObjectToSerial)
	if object == nil {
		return fmt.Errorf("Object not found: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
	}
	buyer, ok := object.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
	}
	// Find seller.
	object = game.Module().Object(req.Sell.ObjectToID, req.Sell.ObjectToSerial)
	if object == nil {
		return fmt.Errorf("Object not found: %s %s", req.Sell.ObjectToID,
			req.Sell.ObjectToSerial)
	}
	seller, ok := object.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", req.Sell.ObjectToID,
			req.Sell.ObjectToSerial)
	}
	// Trade items.
	err := transferItems(seller, buyer, req.Buy.Items)
	if err != nil {
		return fmt.Errorf("Unable to transfer items to buy: %v", err)
	}
	err = transferItems(buyer, seller, req.Sell.Items)
	if err != nil {
		return fmt.Errorf("Unable to transfer items to sell: %v", err)
	}
	return nil
}
