/*
 * request.go
 *
 * Copyright (C) 2020-2025 Dariusz Sikora <ds@isangeles.dev>
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
	"path/filepath"
	"time"

	"github.com/isangeles/flame/area"
	"github.com/isangeles/flame/character"
	flamedata "github.com/isangeles/flame/data"
	"github.com/isangeles/flame/data/res"
	"github.com/isangeles/flame/dialog"
	"github.com/isangeles/flame/effect"
	"github.com/isangeles/flame/item"
	"github.com/isangeles/flame/objects"
	"github.com/isangeles/flame/serial"
	"github.com/isangeles/flame/training"
	"github.com/isangeles/flame/useaction"

	"github.com/isangeles/burn"
	"github.com/isangeles/burn/syntax"

	"github.com/isangeles/fire/config"
	"github.com/isangeles/fire/data"
	"github.com/isangeles/fire/request"
	"github.com/isangeles/fire/response"
)

// handleRequest handles specified client request.
func handleRequest(req clientRequest) {
	resp := response.Response{}
	for _, l := range req.Login {
		err := handleLoginRequest(req.Client, l)
		if err != nil {
			err := fmt.Sprintf("Unable to handle login request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
		// Add user characters.
		game.UpdateUserChars(req.Client.User())
	}
	if req.Client.User() == nil {
		// Request login.
		log.Printf("Authorization requested: %s", req.Client.RemoteAddr())
		resp.Logon = true
		resp.Error = append(resp.Error, fmt.Sprintf("Unauthorized client"))
		req.Client.Out <- resp
		return
	}
	for _, r := range req.NewChar {
		err := handleNewCharRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle new-char request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
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
	for _, de := range req.DialogEnd {
		err := handleDialogEndRequest(req.Client, de)
		if err != nil {
			err := fmt.Sprintf("Unable to handle dialog-end request: %v", err)
			resp.Error = append(resp.Error, err)
			continue
		}
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
	for _, ti := range req.ThrowItems {
		err := handleThrowItemsRequest(req.Client, ti)
		if err != nil {
			err := fmt.Sprintf("Unable to handle throw-items request: %v", err)
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
	for _, r := range req.Equip {
		err := handleEquipRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle equip request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Unequip {
		err := handleUnequipRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle unequip request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Training {
		err := handleTrainingRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle training request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Target {
		err := handleTargetRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle target request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Chat {
		err := handleChatRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle chat request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	for _, r := range req.Save {
		err := handleSaveRequest(req.Client, r)
		if err != nil {
			err := fmt.Sprintf("Unable to handle save request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	if len(req.Load) > 0 {
		err := handleLoadRequest(req.Client, req.Load)
		if err != nil {
			err := fmt.Sprintf("Unable to handle load request: %v", err)
			resp.Error = append(resp.Error, err)
		}
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
	for _, a := range req.Accept {
		handleAcceptRequest(req.Client, a)
	}
	if req.Client.User().Admin() {
		game.pause = req.Pause
	}
	if req.Close > 0 {
		err := handleCloseRequest(req.Client, req.Close)
		if err != nil {
			err := fmt.Sprintf("Unable to handle close request: %v", err)
			resp.Error = append(resp.Error, err)
		}
	}
	updateClient(req.Client, resp)
}

// handleLoginReqest handles login request.
func handleLoginRequest(cli *Client, req request.Login) error {
	user := data.User(req.ID)
	if user == nil || user.Pass() != req.Pass {
		return fmt.Errorf("Invalid ID/password")
	}
	if user.Logged {
		return fmt.Errorf("Already logged")
	}
	game.ActivateUserChars(user)
	cli.SetUser(user)
	return nil
}

// handleNewCharRequest handles new character request.
func handleNewCharRequest(cli *Client, req request.NewChar) error {
	if !game.ValidNewCharacter(req.Data) {
		return fmt.Errorf("Invalid character")
	}
	char := character.New(req.Data)
	game.Chapter().Resources().Characters = append(game.Chapter().Resources().Characters, req.Data)
	err := game.SpawnChar(char)
	if err != nil {
		return fmt.Errorf("Unable to spawn char: %v", err)
	}
	game.AddTranslationAll(res.TranslationData{req.Data.ID, []string{req.Name}})
	cli.User().AddChar(char)
	return nil
}

// handleMoveRequest handles move request.
func handleMoveRequest(cli *Client, req request.Move) error {
	// Retrieve object.
	chapter := game.Chapter()
	ob := chapter.AreaObject(req.ID, req.Serial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ID, req.Serial)
	}
	// Check if object is under client control.
	if !cli.User().Controls(ob.ID(), ob.Serial()) {
		return fmt.Errorf("Object not controled: %s %s", req.ID, req.Serial)
	}
	// Set position.
	char, ok := ob.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", req.ID, req.Serial)
	}
	char.SetDestPoint(req.PosX, req.PosY)
	return nil
}

// handleDialogRequest handles dialog request.
func handleDialogRequest(cli *Client, req request.Dialog) (resp res.ObjectDialogData, err error) {
	// Check if client controls dialog target.
	if !cli.User().Controls(req.TargetID, req.TargetSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Object(req.OwnerID, req.OwnerSerial)
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
	object = game.Object(req.TargetID, req.TargetSerial)
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
		err = fmt.Errorf("Dialog already in progress: %s", req.DialogID)
		return
	}
	dialog.Restart()
	// Set dialog target.
	dialog.SetTarget(target)
	// Make response for the client.
	resp = res.ObjectDialogData{dialog.ID(), dialog.Stage().ID()}
	return
}

// handleDialogAnswerRequest handles dialog answer request.
func handleDialogAnswerRequest(cli *Client, req request.DialogAnswer) (resp res.ObjectDialogData, err error) {
	// Check if client controls dialog target.
	if !cli.User().Controls(req.Dialog.TargetID, req.Dialog.TargetSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
		return
	}
	// Retrieve dialog onwer & target.
	object := game.Object(req.OwnerID, req.OwnerSerial)
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
	object = game.Object(req.TargetID, req.TargetSerial)
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
	if reqDialog.Stage() == nil {
		err = fmt.Errorf("No suitable dialog phase found")
		return
	}
	// Make response for the client.
	resp = res.ObjectDialogData{reqDialog.ID(), reqDialog.Stage().ID()}
	return
}

// handleDialogEndRequest handles dialog end request.
func handleDialogEndRequest(cli *Client, req request.DialogEnd) error {
	// Check if client controls dialog target
	if !cli.User().Controls(req.TargetID, req.TargetSerial) {
		return fmt.Errorf("Object not controlled: %s %s", req.TargetID,
			req.TargetSerial)
	}
	// Retrieve dialog onwer & target
	object := game.Object(req.OwnerID, req.OwnerSerial)
	if object == nil {
		return fmt.Errorf("Dialog owner not found: %s %s", req.OwnerID,
			req.OwnerSerial)
	}
	owner, ok := object.(dialog.Talker)
	if !ok {
		return fmt.Errorf("Invalid dialog onwer: %s %s", req.OwnerID,
			req.OwnerSerial)
	}
	object = game.Object(req.TargetID, req.TargetSerial)
	if object == nil {
		return fmt.Errorf("Dialog target not found: %s %s", req.TargetID,
			req.TargetSerial)
	}
	target, ok := object.(dialog.Talker)
	if !ok {
		return fmt.Errorf("Invalid dialog target: %s %s", req.TargetID,
			req.TargetSerial)
	}
	// Check range
	if !inRange(owner, target) {
		return fmt.Errorf("Objects are not in the minimal range")
	}
	// Retrieve requested dialog from owner
	var dialog *dialog.Dialog
	for _, d := range owner.Dialogs() {
		if d.ID() == req.DialogID {
			dialog = d
		}
	}
	if dialog == nil {
		return fmt.Errorf("Dialog not found: %s", req.DialogID)
	}
	if dialog.Target().ID() != req.TargetID || dialog.Target().Serial() != req.TargetSerial {
		return fmt.Errorf("Dialog not started by the target object: %s", req.DialogID)
	}
	// End dialog
	dialog.Restart()
	dialog.SetTarget(nil)
	return nil
}

// handleTradeRequest handles trade request.
func handleTradeRequest(cli *Client, req request.Trade) (resp response.Trade, err error) {
	// Check if client controls buyer.
	if !cli.User().Controls(req.Buy.ObjectToID, req.Buy.ObjectToSerial) {
		err = fmt.Errorf("Object not controlled: %s %s", req.Buy.ObjectToID,
			req.Buy.ObjectToSerial)
		return
	}
	// Find seller & buyer.
	object := game.Object(req.Sell.ObjectToID, req.Sell.ObjectToSerial)
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
	object = game.Object(req.Buy.ObjectToID, req.Buy.ObjectToSerial)
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
func handleTransferItemsRequest(cli *Client, req request.TransferItems) error {
	// Retrive objects 'to' and 'from'.
	ob := game.Object(req.ObjectToID, req.ObjectToSerial)
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
	ob = game.Object(req.ObjectFromID, req.ObjectFromSerial)
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

// handleThrowItemRequest handles throw items request.
func handleThrowItemsRequest(cli *Client, req request.ThrowItems) error {
	// Retrive object.
	ob := game.Object(req.ObjectID, req.ObjectSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	container, ok := ob.(item.Container)
	if !ok {
		return fmt.Errorf("Object is not a container: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	if !cli.User().Controls(container.ID(), container.Serial()) {
		return fmt.Errorf("Object is not controlled: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	// Remove items.
	switch container := container.(type) {
	case *character.Character:
		if !cli.User().Controls(container.ID(), container.Serial()) && container.Live() {
			return fmt.Errorf("Can't transfer items from: %s %s", req.ObjectID,
				req.ObjectSerial)
		}
		err := removeItems(container, req.Items)
		if err != nil {
			return fmt.Errorf("Unable to remove items: %v", err)
		}
	default:
		return fmt.Errorf("Unsupported object: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	return nil
}

// handleTrainingRequest handles training request.
func handleTrainingRequest(cli *Client, req request.Training) error {
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
	// Retrieve trainer.
	ob = serial.Object(req.TrainerID, req.TrainerSerial)
	if ob == nil {
		return fmt.Errorf("Trainer object not found: %s %s", req.TrainerID,
			req.TrainerSerial)
	}
	trainer, ok := ob.(training.Trainer)
	if !ok {
		return fmt.Errorf("Trainer object is not a trainer: %s %s", req.TrainerID,
			req.TrainerSerial)
	}
	// Retrive training.
	var train *training.TrainerTraining
	for _, t := range trainer.Trainings() {
		if t.ID() == req.TrainingID {
			train = t
			break
		}
	}
	if train == nil {
		return fmt.Errorf("Training not found: %s", req.TrainingID)
	}
	// Check range.
	if !inRange(user, trainer) {
		return fmt.Errorf("Objects are not in the minimal range")
	}
	user.Use(train)
	return nil
}

// handleUseRequest handles use request.
func handleUseRequest(cli *Client, req request.Use) error {
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
	}
	// Check range.
	if !inRange(user, ob) {
		return fmt.Errorf("Objects are not in the minimal range")
	}
	// Use object.
	err := user.Use(usable)
	if err != nil {
		return fmt.Errorf("Unable to use object: %v", err)
	}
	// Notify near chars.
	useResp := response.Use{
		ObjectID:     req.ObjectID,
		ObjectSerial: req.ObjectSerial,
		UserID:       req.UserID,
		UserSerial:   req.UserSerial,
	}
	resp := response.Response{Use: []response.Use{useResp}}
	go game.NotifyNearObjects(user, resp)
	return nil
}

// handleEquipRequest handles equip request.
func handleEquipRequest(cli *Client, req request.Equip) error {
	// Retrieve object.
	ob := serial.Object(req.CharID, req.CharSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.CharID,
			req.CharSerial)
	}
	if !cli.User().Controls(req.CharID, req.CharSerial) {
		return fmt.Errorf("Object is not controled: %s %s", req.CharID,
			req.CharSerial)
	}
	object, ok := ob.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", req.CharID,
			req.CharSerial)
	}
	// Retrieve item.
	it := object.Inventory().Item(req.ItemID, req.ItemSerial)
	if it == nil {
		return fmt.Errorf("Item not found in object inventory: %s %s",
			req.ItemID, req.ItemSerial)
	}
	// Equip item.
	eqItem, ok := it.Item.(item.Equiper)
	if !ok {
		return fmt.Errorf("Item is not equipable: %s %s", it.ID(),
			it.Serial())
	}
	if object.Equipment().Equiped(eqItem) {
		return fmt.Errorf("Item is already equiped: %s %s", it.ID(),
			it.Serial())
	}
	err := equip(object.Equipment(), eqItem, req.Slots)
	if err != nil {
		return fmt.Errorf("Unable to equip item in slot: %v", err)
	}
	return nil
}

// handleUnequipRequest handles unequip request.
func handleUnequipRequest(cli *Client, req request.Unequip) error {
	// Retrieve object.
	ob := serial.Object(req.CharID, req.CharSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.CharID,
			req.CharSerial)
	}
	if !cli.User().Controls(req.CharID, req.CharSerial) {
		return fmt.Errorf("Object is not controled: %s %s", req.CharID,
			req.CharSerial)
	}
	object, ok := ob.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", req.CharID,
			req.CharSerial)
	}
	// Retrieve item.
	it := object.Inventory().Item(req.ItemID, req.ItemSerial)
	if it == nil {
		return fmt.Errorf("Item not found in object inventory: %s %s",
			req.ItemID, req.ItemSerial)
	}
	// Equip item.
	eqItem, ok := it.Item.(item.Equiper)
	if !ok {
		return fmt.Errorf("Item is not equipable: %s %s", req.ItemID,
			req.ItemSerial)
	}
	object.Equipment().Unequip(eqItem)
	return nil
}

// handleChatRequest handles chat request.
func handleChatRequest(cli *Client, req request.Chat) error {
	// Retrieve object.
	ob := serial.Object(req.ObjectID, req.ObjectSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	if !cli.User().Controls(req.ObjectID, req.ObjectSerial) {
		return fmt.Errorf("Object is not controled: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	logger, ok := ob.(objects.Logger)
	if !ok {
		return fmt.Errorf("Object is has no chat log: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	msg := objects.NewMessage(req.Message, req.Translated)
	logger.ChatLog().Add(msg)
	areaOb, ok := logger.(area.Object)
	if !ok {
		return nil
	}
	// Notify near chars.
	chatResp := response.Chat{
		ObjectID:     req.ObjectID,
		ObjectSerial: req.ObjectSerial,
		Message:      req.Message,
		Translated:   req.Translated,
		Time:         msg.Time,
	}
	resp := response.Response{Chat: []response.Chat{chatResp}}
	go game.NotifyNearObjects(areaOb, resp)
	return nil
}

// handleTargetRequest handles target request.
func handleTargetRequest(cli *Client, req request.Target) error {
	// Retrieve object.
	ob := serial.Object(req.ObjectID, req.ObjectSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.ObjectID,
			req.ObjectSerial)
	}
	if !cli.User().Controls(req.ObjectID, req.ObjectSerial) {
		return fmt.Errorf("Object is not controled: %s %s", ob.ID(),
			ob.Serial())
	}
	char, ok := ob.(*character.Character)
	if !ok {
		return fmt.Errorf("Object is not a character: %s %s", ob.ID(),
			ob.Serial())
	}
	// Retrieve target.
	if len(req.TargetID+req.TargetSerial) < 1 {
		char.SetTarget(nil)
		return nil
	}
	ob = serial.Object(req.TargetID, req.TargetSerial)
	if ob == nil {
		return fmt.Errorf("Object not found: %s %s", req.TargetID,
			req.TargetSerial)
	}
	tar, ok := ob.(effect.Target)
	if !ok {
		return fmt.Errorf("Object is not targetable: %s %s", ob.ID(),
			ob.Serial())
	}
	// Set target.
	char.SetTarget(tar)
	return nil
}

// handleSaveRequest handles save request.
func handleSaveRequest(cli *Client, saveName string) error {
	if !cli.User().Admin() {
		return fmt.Errorf("You are not the admin")
	}
	path := filepath.Join(config.ModulesPath, saveName)
	err := flamedata.ExportModule(path, game.Data())
	if err != nil {
		return fmt.Errorf("Unable to export module file: %v", err)
	}
	return nil
}

// handleLoadRequest handles load request.
func handleLoadRequest(cli *Client, saveName string) error {
	if !cli.User().Admin() {
		return fmt.Errorf("You are not the admin")
	}
	// Import module.
	path := filepath.Join(config.ModulesPath, saveName+flamedata.ModuleFileExt)
	data, err := flamedata.ImportModule(path)
	if err != nil {
		return fmt.Errorf("Unable to import module file: %v", err)
	}
	// Send load data on load channel.
	loadResp := response.Load{saveName, data}
	loadGame := func() { load <- loadResp }
	go loadGame()
	return nil
}

// handleCommandRequest handles command request.
func handleCommandRequest(cli *Client, cmdText string) (resp response.Command, err error) {
	if !cli.User().Admin() {
		err = fmt.Errorf("You are not the admin")
		return
	}
	exp, err := syntax.NewSTDExpression(cmdText)
	if err != nil {
		err = fmt.Errorf("Invalid command syntax: %v", err)
		return
	}
	res, out := burn.HandleExpression(exp)
	resp = response.Command{res, out}
	return
}

// handleAcceptRequest handles accept request.
func handleAcceptRequest(cli *Client, id int) {
	confirm := clientConfirm{id, cli}
	confirmReq := func() { confirmed <- &confirm }
	go confirmReq()
}

// handleCloseRequest handles close request.
func handleCloseRequest(cli *Client, timeNano int64) error {
	if !cli.User().Admin() {
		return fmt.Errorf("You are not the admin")
	}
	closeTime := time.Unix(0, timeNano)
	closeFunc := func() { close = true }
	log.Printf("Server going down at: %v", closeTime)
	time.AfterFunc(time.Until(closeTime), closeFunc)
	return nil
}
