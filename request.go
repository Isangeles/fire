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
	"github.com/isangeles/flame/module/object"
	"github.com/isangeles/flame/module/objects"
	"github.com/isangeles/flame/module/serial"
	"github.com/isangeles/flame/module/skill"
	"github.com/isangeles/flame/module/useaction"

	"github.com/isangeles/burn"
	"github.com/isangeles/burn/syntax"

	"github.com/isangeles/fire/client"
	"github.com/isangeles/fire/data"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/response"
	"github.com/isangeles/fire/user"
)

// handleRequest handles specified client request.
func handleRequest(req clientRequest) {
	resp := response.Response{}
	for _, l := range req.Login {
		err := handleLoginRequest(req.Client, l)
		if err != nil {
			err := fmt.Sprintf("Unable to handle login request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	if req.Client.User() == nil {
		// Request login.
		log.Printf("Authorization requested: %s", req.Client.RemoteAddr())
		resp.Logon = true
		req.Client.Out <- resp
		return
	}
	for _, nc := range req.NewChar {
		r, err := handleNewCharRequest(req.Client, nc)
		if err != nil {
			err := fmt.Sprintf("Unable to handle new-char request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		resp.NewChar = append(resp.NewChar, r)
	}
	for _, m := range req.Move {
		err := handleMoveRequest(req.Client, m)
		if err != nil {
			err := fmt.Sprintf("Unable to handle move request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, d := range req.Dialog {
		r, err := handleDialogRequest(req.Client, d)
		if err != nil {
			err := fmt.Sprintf("Unable to handle dialog request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		resp.Dialog = append(resp.Dialog, r)
	}
	for _, da := range req.DialogAnswer {
		r, err := handleDialogAnswerRequest(req.Client, da)
		if err != nil {
			err := fmt.Sprintf("Unable to handle dialog-answer request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		resp.Dialog = append(resp.Dialog, r)
	}
	for _, t := range req.Trade {
		r, err := handleTradeRequest(req.Client, t)
		if err != nil {
			err := fmt.Sprintf("Unable to handle trade request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		// Send response to trade target owner.
		charResp := charResponse{
			CharID:     t.Buy.ObjectFromID,
			CharSerial: t.Buy.ObjectFromSerial,
		}
		charResp.Response.Trade = append(charResp.Response.Trade, r)
		sendCharResp := func() { charResponses <- charResp }
		go sendCharResp()
	}
	for _, ti := range req.TransferItems {
		err := handleTransferItemsRequest(req.Client, ti)
		if err != nil {
			err := fmt.Sprintf("Unable to handle transfer-items request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, s := range req.Skill {
		err := handleSkillRequest(req.Client, s)
		if err != nil {
			err := fmt.Sprintf("Unable to handle skill request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Use {
		err := handleUseRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle use request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, a := range req.Accept {
		handleAcceptRequest(req.Client, a)
	}
	for _, c := range req.Command {
		r, err := handleCommandRequest(req.Client, c)
		if err != nil {
			err := fmt.Sprintf("Unable to handle command request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		resp.Command = append(resp.Command, r)
	}
	req.Client.Out <- resp
}

// handleLoginReqest handles login request.
func handleLoginRequest(cli *client.Client, req request.Login) error {
	user := data.User(req.ID)
	if user == nil || user.Pass() != req.Pass {
		return fmt.Errorf("Invalid ID/password")
	}
	if user.Logged {
		return fmt.Errorf("Already logged")
	}
	cli.SetUser(user)
	return nil
}

// handleNewCharRequest handles new character request.
func handleNewCharRequest(cli *client.Client, charData res.CharacterData) (resp res.CharacterData, err error) {
	if !game.ValidNewCharacter(charData) {
		err = fmt.Errorf("Invalid character")
		return
	}
	char, err := game.SpawnChar(charData)
	if err != nil {
		err = fmt.Errorf("Unable to spawn char: %v", err)
		return
	}
	cli.User().Chars = append(cli.User().Chars, user.Character{char.ID(), char.Serial()})
	resp = char.Data()
	return
}

// handleMoveRequest handles move request.
func handleMoveRequest(cli *client.Client, req request.Move) error {
	// Retrieve object.
	chapter := game.Module().Chapter()
	ob := chapter.Object(req.ID, req.Serial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ID, req.Serial)
	}
	// Check if object is under client control.
	if cli.User().Controls(ob.ID(), ob.Serial()) {
		return fmt.Errorf("Object not controled: %s %s", req.ID, req.Serial)
	}
	// Set position.
	posOb, ok := ob.(objects.Positioner)
	if !ok {
		return fmt.Errorf("Object without position: %s %s", req.ID, req.Serial)
	}
	posOb.SetPosition(req.PosX, req.PosY)
	return nil
}

// handleDialogRequest handles dialog request.
func handleDialogRequest(cli *client.Client, req request.Dialog) (resp res.ObjectDialogData, err error) {
	// Check if client controls dialog target.
	if !cli.User().Controls(req.TargetID, req.TargetSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Module().Object(req.OwnerID, req.OwnerSerial)
	if object == nil {
		err = fmt.Errorf("Dialog owner not found: %s %s", req.OwnerID,
			req.OwnerSerial)
		return
	}
	owner, ok := object.(dialog.Talker)
	if !ok {
		err = fmt.Errorf("Invalid dialog onwer: %s %s", req.OwnerID,
			req.OwnerSerial)
		return
	}
	object = game.Module().Object(req.TargetID, req.TargetSerial)
	if object == nil {
		err = fmt.Errorf("Dialog target not found: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	target, ok := object.(dialog.Talker)
	if !ok {
		err = fmt.Errorf("Invalid dialog target: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Check range.
	if !inRange(owner, target) {
		err = fmt.Errorf("Objects are not in the minimal range")
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
		err = fmt.Errorf("Dialog not found: %s", req.DialogID)
		return
	}
	if dialog.Target() != nil {
		err = fmt.Errorf("Dialog already started")
		return
	}
	// Set dialog target.
	dialog.SetTarget(target)
	// Make response for the client.
	resp = res.ObjectDialogData{dialog.ID(), dialog.Stage().ID()}
	return
}

// handleDialogAnswerRequest handles dialog answer request.
func handleDialogAnswerRequest(cli *client.Client, req request.DialogAnswer) (resp res.ObjectDialogData, err error) {
	// Check if client controls dialog target.
	if !cli.User().Controls(req.Dialog.TargetID, req.Dialog.TargetSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Module().Object(req.OwnerID, req.OwnerSerial)
	if object == nil {
		err = fmt.Errorf("Dialog owner not found: %s %s", req.OwnerID,
			req.OwnerSerial)
		return
	}
	owner, ok := object.(dialog.Talker)
	if !ok {
		err = fmt.Errorf("Invalid dialog onwer: %s %s", req.OwnerID,
			req.OwnerSerial)
		return
	}
	object = game.Module().Object(req.TargetID, req.TargetSerial)
	if object == nil {
		err = fmt.Errorf("Dialog target not found: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	target, ok := object.(dialog.Talker)
	if !ok {
		err = fmt.Errorf("Invalid dialog target: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Check range.
	if !inRange(owner, target) {
		err = fmt.Errorf("Objects are not in the minimal range")
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
		err = fmt.Errorf("Dialog not found: %s", req.DialogID)
		return
	}
	// Check dialog target.
	if reqDialog.Target() == nil {
		err = fmt.Errorf("Dialog not started: %s", reqDialog.ID())
		return
	}
	if reqDialog.Target().ID() != req.TargetID ||
		reqDialog.Target().Serial() != req.TargetSerial {
		err = fmt.Errorf("Target different then specified in request: %s %s",
			reqDialog.Target().ID(), reqDialog.Target().Serial())
		return
	}
	// Apply answer.
	if reqDialog.Stage() == nil {
		err = fmt.Errorf("Requested dialog has no active stage: %s",
			reqDialog.ID())
		return
	}
	var answer *dialog.Answer
	for _, a := range reqDialog.Stage().Answers() {
		if a.ID() == req.AnswerID {
			answer = a
		}
	}
	if answer == nil {
		err = fmt.Errorf("Requested answer not found: %s", req.AnswerID)
		return
	}
	reqDialog.Next(answer)
	// Make response for the client.
	resp = res.ObjectDialogData{reqDialog.ID(), reqDialog.Stage().ID()}
	return
}

// handleTradeRequest handles trade request.
func handleTradeRequest(cli *client.Client, req request.Trade) (resp response.Trade, err error) {
	// Check if client controls buyer.
	if !cli.User().Controls(req.Buy.ObjectToID, req.Buy.ObjectToSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
		return
	}
	// Find seller & buyer.
	object := game.Module().Object(req.Sell.ObjectToID, req.Sell.ObjectToSerial)
	if object == nil {
		err = fmt.Errorf("Seller not found: %s %s", req.Sell.ObjectToID,
			req.Sell.ObjectToSerial)
		return
	}
	seller, ok := object.(*character.Character)
	if !ok {
		err = fmt.Errorf("Seller is not a character: %s %s", req.Sell.ObjectToID,
			req.Sell.ObjectToSerial)
		return
	}
	object = game.Module().Object(req.Buy.ObjectToID, req.Buy.ObjectToSerial)
	if object == nil {
		err = fmt.Errorf("Buyer not found: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
		return
	}
	buyer, ok := object.(*character.Character)
	if !ok {
		err = fmt.Errorf("Buyer is not a character: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
		return
	}
	// Check range.
	if !inRange(buyer, seller) {
		err = fmt.Errorf("Objects are not in the minimal range")
		return
	}
	// Send confiramtion request to seller owner.
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
	resp = response.Trade{
		ID:           confirmReq.ID,
		BuyerID:      req.Buy.ObjectToID,
		BuyerSerial:  req.Buy.ObjectToSerial,
		SellerID:     req.Sell.ObjectToID,
		SellerSerial: req.Sell.ObjectToSerial,
		ItemsBuy:     req.Buy.Items,
		ItemsSell:    req.Sell.Items,
	}
	return
}

// handleTransferItemsRequest handles transfer request.
func handleTransferItemsRequest(cli *client.Client, req request.TransferItems) error {
	// Retrive objects 'to' and 'from'.
	ob := game.Module().Object(req.ObjectToID, req.ObjectToSerial)
	if ob == nil {
		return fmt.Errorf("Object 'to' not found: %s %s", req.ObjectToID,
			req.ObjectToSerial)
	}
	to, ok := ob.(item.Container)
	if !ok {
		return fmt.Errorf("Object 'to' is not a container: %s %s", req.ObjectToID,
			req.ObjectToSerial)
	}
	if !cli.User().Controls(to.ID(), to.Serial()) {
		return fmt.Errorf("Object 'to' is not controlled: %s %s", req.ObjectToID,
			req.ObjectToSerial)
	}
	ob = game.Module().Object(req.ObjectFromID, req.ObjectFromSerial)
	if ob == nil {
		return fmt.Errorf("Object 'from' not found: %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
	}
	from, ok := ob.(item.Container)
	if !ok {
		return fmt.Errorf("Object 'from' is not a container: %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
	}
	// Check range.
	if !inRange(from, to) {
		return fmt.Errorf("Objects are not in the minimal range")
	}
	// Transfer items.
	switch from := from.(type) {
	case *character.Character:
		if !cli.User().Controls(from.ID(), from.Serial()) && from.Live() {
			return fmt.Errorf("Can't transfer items from: %s %s", req.ObjectFromID,
				req.ObjectFromSerial)
		}
		log.Printf("items: %v", from.Inventory().Items())
		err := transferItems(from, to, req.Items)
		if err != nil {
			return fmt.Errorf("Unable to transfer items: %v", err)
		}
	default:
		return fmt.Errorf("Unsupported object 'from': %s %s", req.ObjectFromID,
			req.ObjectFromSerial)
	}
	return nil
}

// handleSkillRequest handles skill request.
func handleSkillRequest(cli *client.Client, req request.Skill) error {
	// Retrieve object.
	ob := game.Module().Object(req.ObjectID, req.ObjectSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	user, ok := ob.(skill.User)
	if !ok {
		return fmt.Errorf("Object is not a skill user: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	// Retrieve skill.
	var skill *skill.Skill
	for _, s := range user.Skills() {
		if s.ID() == req.SkillID {
			skill = s
		}
	}
	if skill == nil {
		return fmt.Errorf("Skill not found: %s", req.SkillID)
	}
	// Use skill.
	user.Use(skill)
	return nil
}

// handleUseRequest handles use request.
func handleUseRequest(cli *client.Client, req request.Use) error {
	// Retrieve user.
	ob := serial.Object(req.UserID, req.UserSerial)
	if ob == nil {
		return fmt.Errorf("User not found: %s %s", req.UserID, req.UserSerial)
	}
	if !cli.User().Controls(req.UserID, req.UserSerial) {
		return fmt.Errorf("User is not controled: %s %s", req.UserID,
			req.UserSerial)
	}
	user, ok := ob.(*character.Character)
	if !ok {
		return fmt.Errorf("User is not a character: %s %s", req.UserID,
			req.UserSerial)
	}
	// Retrieve usable object.
	usable := charSkillRecipe(user, req.ObjectID)
	if usable == nil {
		// Search for item or area object.
		ob = serial.Object(req.ObjectID, req.ObjectSerial)
		if ob == nil {
			return fmt.Errorf("Object not found: %s %s", req.ObjectID,
				req.ObjectSerial)
		}
		u, ok := ob.(useaction.Usable)
		if !ok {
			return fmt.Errorf("Object is not usable: %s %s", req.ObjectID,
				req.ObjectSerial)
		}
		usable = u
	}
	// Check if usable object can be used.
	switch usable := usable.(type) {
	case item.Item:
		if user.Inventory().Item(usable.ID(), usable.Serial()) == nil {
			return fmt.Errorf("User doesn't own usable item: %s %s",
				usable.ID(), usable.Serial())
		}
	case *object.Object:
		if !inRange(user, usable) {
			return fmt.Errorf("Objects are not in the minimal range")
		}
	}
	// Use object.
	user.Use(usable)
	return nil
}

// handleAcceptRequest handles accept request.
func handleAcceptRequest(cli *client.Client, id int) {
	confirm := clientConfirm{id, cli}
	confirmReq := func() { confirmed <- &confirm }
	go confirmReq()
}

// handleCommandRequest handles command request.
func handleCommandRequest(cli *client.Client, cmdText string) (resp response.Command, err error) {
	exp, err := syntax.NewSTDExpression(cmdText)
	if err != nil {
		err = fmt.Errorf("Invalid command syntax: %v", err)
		return
	}
	res, out := burn.HandleExpression(exp)
	resp = response.Command{res, out}
	return
}
