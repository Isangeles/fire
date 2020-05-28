/*
 * request.go
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
	"log"

	"github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/module/character"
	"github.com/isangeles/flame/module/dialog"
	"github.com/isangeles/flame/module/item"
	"github.com/isangeles/flame/module/objects"

	"github.com/isangeles/burn"
	"github.com/isangeles/burn/syntax"

	"github.com/isangeles/fire/client"
	"github.com/isangeles/fire/data"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/response"
)

// handleRequest handles specified client request.
func handleRequest(req clientRequest) {
	resp := response.Response{}
	for _, l := range req.Login {
		handleLoginRequest(req.Client, l, &resp)
	}
	if req.Client.User() == nil {
		// Request login.
		log.Printf("Authorization requested: %s",
			req.Client.RemoteAddr())
		resp.Logon = true
		req.Client.Out <- resp
		return
	}
	for _, nc := range req.NewChar {
		handleNewCharRequest(req.Client, nc, &resp)
	}
	for _, m := range req.Move {
		handleMoveRequest(req.Client, m, &resp)
	}
	for _, d := range req.Dialog {
		handleDialogRequest(req.Client, d, &resp)
	}
	for _, da := range req.DialogAnswer {
		handleDialogAnswerRequest(req.Client, da, &resp)
	}
	for _, t := range req.Trade {
		handleTradeRequest(req.Client, t, &resp)
	}
	for _, ti := range req.TransferItems {
		handleTransferItemsRequest(req.Client, ti, &resp)
	}
	for _, a := range req.Accept {
		handleAcceptRequest(req.Client, a, &resp)
	}
	for _, c := range req.Command {
		handleCommandRequest(req.Client, c, &resp)
	}
	req.Client.Out <- resp
}

// handleLoginReqest handles login request.
func handleLoginRequest(cli *client.Client, req request.Login, resp *response.Response) {
	user := data.User(req.ID)
	if user == nil || user.Pass() != req.Pass {
		err := fmt.Sprintf("Invalid ID/password")
		resp.Errors = append(resp.Errors, err)
		return
	}
	if user.Logged {
		err := fmt.Sprintf("Already logged")
		resp.Errors = append(resp.Errors, err)
		return
	}
	cli.SetUser(user)
	resp.Logon = false
}

// handleNewCharRequest handles new character request.
func handleNewCharRequest(cli *client.Client, charData res.CharacterData, resp *response.Response) {
	char, err := game.SpawnChar(charData)
	if err != nil {
		log.Printf("handle new char: unable to spawn char: %v",
			err)
		resp.Errors = append(resp.Errors, "Internal error")
	}
	cli.User().Chars = append(cli.User().Chars, char.ID()+char.Serial())
	resp.NewChars = append(resp.NewChars, char.Data())
}

// handleMoveRequest handles move request.
func handleMoveRequest(cli *client.Client, req request.Move, resp *response.Response) {
	// Retrieve object.
	chapter := game.Module().Chapter()
	ob := chapter.Object(req.ID, req.Serial)
	if ob == nil {
		resp.Errors = append(resp.Errors, "Object not found")
		return
	}
	// Check if object is under client control.
	control := false
	for _, c := range cli.User().Chars {
		if c != ob.ID()+ob.Serial() {
			continue
		}
		control = true
		break
	}
	if !control {
		resp.Errors = append(resp.Errors, "Object not controled")
		return
	}
	// Set position.
	posOb, ok := ob.(objects.Positioner)
	if !ok {
		resp.Errors = append(resp.Errors, "Object without position")
		return
	}
	posOb.SetPosition(req.PosX, req.PosY)
}

// handleDialogRequest handles dialog request.
func handleDialogRequest(cli *client.Client, req request.Dialog, resp *response.Response) {
	// Check if client controls dialog target.
	if !cli.OwnsChar(req.TargetID, req.TargetSerial) {
		err := fmt.Sprintf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Module().Object(req.OwnerID, req.OwnerSerial)
	if object == nil {
		err := fmt.Sprintf("Dialog owner not found: %s %s", req.OwnerID,
			req.OwnerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	owner, ok := object.(dialog.Talker)
	if !ok {
		err := fmt.Sprintf("Invalid dialog onwer: %s %s", req.OwnerID,
			req.OwnerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	object = game.Module().Object(req.TargetID, req.TargetSerial)
	if object == nil {
		err := fmt.Sprintf("Dialog target not found: %s %s", req.TargetID,
			req.TargetSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	target, ok := object.(dialog.Talker)
	if !ok {
		err := fmt.Sprintf("Invalid dialog target: %s %s", req.TargetID,
			req.TargetSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Retrieve requested dialog from owner.
	var dialog *dialog.Dialog
	for _, d := range owner.Dialogs() {
		if d.ID() == req.DialogID {
			dialog = d
		}
	}
	if dialog == nil {
		err := fmt.Sprintf("Dialog not found: %s", req.DialogID)
		resp.Errors = append(resp.Errors, err)
		return
	}
	if dialog.Target() != nil {
		err := fmt.Sprintf("Dialog already started")
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Set dialog target.
	dialog.SetTarget(target)
	// Make response for the client.
	dialogResp := res.ObjectDialogData{dialog.ID(), dialog.Stage().ID()}
	resp.Dialog = append(resp.Dialog, dialogResp)
}

// handleDialogAnswerRequest handles dialog answer request.
func handleDialogAnswerRequest(cli *client.Client, req request.DialogAnswer, resp *response.Response) {
	// Check if client controls dialog target.
	if !cli.OwnsChar(req.Dialog.TargetID, req.Dialog.TargetSerial) {
		err := fmt.Sprintf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Module().Object(req.OwnerID, req.OwnerSerial)
	if object == nil {
		err := fmt.Sprintf("Dialog owner not found: %s %s", req.OwnerID,
			req.OwnerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	owner, ok := object.(dialog.Talker)
	if !ok {
		err := fmt.Sprintf("Invalid dialog onwer: %s %s", req.OwnerID,
			req.OwnerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Retrieve requested dialog from owner.
	var reqDialog *dialog.Dialog
	for _, d := range owner.Dialogs() {
		if d.ID() == req.DialogID {
			reqDialog = d
		}
	}
	if reqDialog == nil {
		err := fmt.Sprintf("Dialog not found: %s", req.DialogID)
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Check dialog target.
	if reqDialog.Target() == nil {
		err := fmt.Sprintf("Dialog not started")
		resp.Errors = append(resp.Errors, err)
		return
	}
	if reqDialog.Target().ID() != req.TargetID ||
		reqDialog.Target().Serial() != req.TargetSerial {
		err := fmt.Sprintf("Target different then specified in request")
		resp.Errors = append(resp.Errors, err)
		return
	}
	// Apply answer.
	if reqDialog.Stage() == nil {
		err := fmt.Sprintf("Requested dialog has no active stage")
		resp.Errors = append(resp.Errors, err)
		return
	}
	var answer *dialog.Answer
	for _, a := range reqDialog.Stage().Answers() {
		if a.ID() == req.AnswerID {
			answer = a
		}
	}
	if answer == nil {
		err := fmt.Sprintf("Requested answer not found: %s", req.AnswerID)
		resp.Errors = append(resp.Errors, err)
		return
	}
	reqDialog.Next(answer)
	// Make response for the client.
	dialogResp := res.ObjectDialogData{reqDialog.ID(), reqDialog.Stage().ID()}
	resp.Dialog = append(resp.Dialog, dialogResp)
}

// handleTradeRequest handles trade request.
func handleTradeRequest(cli *client.Client, req request.Trade, resp *response.Response) {
	if !cli.OwnsChar(req.BuyerID, req.BuyerSerial) {
		err := fmt.Sprintf("Object not controlled: %s %s", req.BuyerID,
			req.BuyerSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	object := game.Module().Object(req.SellerID, req.SellerSerial)
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
	confirmReq := charConfirmRequest{
		clientRequest: clientRequest{
			Request: &request.Request{Trade: []request.Trade{req}},
			Client:  cli,
		},
		CharID:     seller.ID(),
		CharSerial: seller.Serial(),
		ID:         len(pendingReqs),
	}
	addConfirmReq := func() { confirmRequests <- confirmReq }
	go addConfirmReq()
	tradeResp := response.Trade{
		ID:           confirmReq.ID,
		BuyerID:      req.BuyerID,
		BuyerSerial:  req.BuyerSerial,
		SellerID:     req.SellerID,
		SellerSerial: req.SellerSerial,
		ItemsBuy:     req.ItemsBuy,
		ItemsSell:    req.ItemsSell,
	}
	charResp := charResponse{CharID: seller.ID(), CharSerial: seller.Serial()}
	charResp.Response.Trade = append(charResp.Response.Trade, tradeResp)
	sendCharResp := func() { charResponses <- charResp }
	go sendCharResp()
}

// handleTransferItemsRequest handles transfer request.
func handleTransferItemsRequest(cli *client.Client, req request.TransferItems, resp *response.Response) {
	ob := game.Module().Object(req.ObjectToID, req.ObjectToSerial)
	if ob == nil {
		err := fmt.Sprintf("Object 'to' not found: %s %s", req.ObjectToID,
			req.ObjectToSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	to, ok := ob.(item.Container)
	if !ok {
		err := fmt.Sprintf("Object 'to' is not a container: %s %s", req.ObjectToID,
			req.ObjectToSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	if !cli.OwnsChar(to.ID(), to.Serial()) {
		err := fmt.Sprintf("Object 'to' is not controlled: %s %s", req.ObjectToID,
			req.ObjectToSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	ob = game.Module().Object(req.ObjectFromID, req.ObjectFromSerial)
	if ob == nil {
		err := fmt.Sprintf("Object 'from' not found: %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	from, ok := ob.(item.Container)
	if !ok {
		err := fmt.Sprintf("Object 'from' is not a container: %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
	switch from := from.(type) {
	case *character.Character:
		if !cli.OwnsChar(from.ID(), from.Serial()) && from.Live() {
			err := fmt.Sprintf("Can't transfer items from: %s %s", req.ObjectFromID,
				req.ObjectFromSerial)
			resp.Errors = append(resp.Errors, err)
			return
		}
		log.Printf("items: %v", from.Inventory().Items())
		err := transferItems(from, to, req.Items)
		if err != nil {
			err := fmt.Sprintf("Unable to transfer items: %v", err)
			resp.Errors = append(resp.Errors, err)
			return
		}
	default:
		err := fmt.Sprintf("Unsupported object 'from': %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
		resp.Errors = append(resp.Errors, err)
		return
	}
}

// handleAcceptRequest handles accept request.
func handleAcceptRequest(cli *client.Client, id int, resp *response.Response) {
	confirm := clientConfirm{id, cli}
	confirmReq := func() { confirmed <- &confirm }
	go confirmReq()
}

// handleCommandRequest handles command request.
func handleCommandRequest(cli *client.Client, cmdText string, resp *response.Response) {
	exp, err := syntax.NewSTDExpression(cmdText)
	if err != nil {
		err := fmt.Sprintf("Invalid command syntax: %v", err)
		resp.Errors = append(resp.Errors, err)
		return
	}
	res, out := burn.HandleExpression(exp)
	resp.Command = append(resp.Command, response.Command{res, out})
}
