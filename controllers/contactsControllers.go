package controllers

import (
	"cig-exchange-libs"
	models "cig-exchange-libs/models"
	"encoding/json"
	"net/http"
)

var CreateContact = func(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("user").(uint) //Grab the id of the user that send the request
	contact := &models.Contact{}

	err := json.NewDecoder(r.Body).Decode(contact)
	if err != nil {
		cigExchange.Respond(w, cigExchange.Message(false, "Error while decoding request body"))
		return
	}

	contact.UserId = user
	resp := contact.Create()
	cigExchange.Respond(w, resp)
}

var GetContactsFor = func(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("user").(uint)
	data := models.GetContacts(id)
	resp := cigExchange.Message(true, "success")
	resp["data"] = data
	cigExchange.Respond(w, resp)
}
