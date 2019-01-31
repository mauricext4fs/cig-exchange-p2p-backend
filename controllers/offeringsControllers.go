package controllers

import (
	"cig-exchange-libs"
	models "cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOffering handles GET api/offerings/{offering_id} endpoint
var GetOffering = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	offeringID := mux.Vars(r)["offering_id"]

	// query offering from db
	offering, err := models.GetOffering(offeringID)
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	cigExchange.Respond(w, offering)
}

// CreateOffering handles POST api/offerings endpoint
var CreateOffering = func(w http.ResponseWriter, r *http.Request) {

	offering := &models.Offering{}
	// decode offering object from request body
	err := json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	// insert offering into db
	err = offering.Create()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	cigExchange.Respond(w, offering)
}

// UpdateOffering handles PATCH api/offerings/{offering_id} endpoint
var UpdateOffering = func(w http.ResponseWriter, r *http.Request) {

	offeringID := mux.Vars(r)["offering_id"]
	offering := &models.Offering{}
	// decode offering object from request body
	err := json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	// set the offering UUID
	offering.ID = offeringID

	// update offering
	err = offering.Update()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	cigExchange.Respond(w, offering)
}

// DeleteOffering handles PATCH api/offerings/{offering_id} endpoint
var DeleteOffering = func(w http.ResponseWriter, r *http.Request) {

	offeringID := mux.Vars(r)["offering_id"]
	offering := &models.Offering{
		ID: offeringID,
	}

	// delete offering
	err := offering.Delete()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	w.WriteHeader(204)
}

// GetOfferings handles GET api/offerings endpoint
var GetOfferings = func(w http.ResponseWriter, r *http.Request) {

	// query all offerings from db
	offerings, err := models.GetOfferings()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	cigExchange.Respond(w, offerings)
}
