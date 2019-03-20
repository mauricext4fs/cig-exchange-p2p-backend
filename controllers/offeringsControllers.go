package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	models "cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOffering handles GET api/offerings/{offering_id} endpoint
var GetOffering = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		cigExchange.RespondWithError(w, 401, err)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
		return
	}

	// query offering from db
	offering, err := models.GetOffering(offeringID)
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	// check if organisation id matches
	if offering.OrganisationID != organisationID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("Offering doesn't exist for the organisation"))
		return
	}
	cigExchange.Respond(w, offering)
}

// CreateOffering handles POST api/offerings endpoint
var CreateOffering = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		cigExchange.RespondWithError(w, 401, err)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
		return
	}

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	if offering.OrganisationID != organisationID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
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

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		cigExchange.RespondWithError(w, 401, err)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
		return
	}

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	if offering.OrganisationID != organisationID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
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

// DeleteOffering handles DELETE api/offerings/{offering_id} endpoint
var DeleteOffering = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// query offering from db first to validate the permissions
	offering, err := models.GetOffering(offeringID)
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	if offering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// delete offering
	apiError := offering.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	w.WriteHeader(204)
}

// GetOfferings handles GET api/offerings endpoint
var GetOfferings = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		cigExchange.RespondWithError(w, 401, err)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		cigExchange.RespondWithError(w, 401, fmt.Errorf("No access rights for the organisation"))
		return
	}

	// query all offerings from db
	offerings, err := models.GetOrganisationOfferings(organisationID)
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}
	cigExchange.Respond(w, offerings)
}

// GetAllOfferings handles GET api/offerings endpoint
// does not perform JWT based organisation filtering
var GetAllOfferings = func(w http.ResponseWriter, r *http.Request) {

	// query all offerings from db
	offerings, err := models.GetOfferings()
	if err != nil {
		cigExchange.RespondWithError(w, 500, err)
		return
	}

	// extended response with organisation and org website
	type offeringsReponse struct {
		*models.Offering
		OrganisationName string `json:"organisation"`
		OrganisationURL  string `json:"organisation_website"`
	}

	// add organisation name to offerings structs
	respOfferings := make([]*offeringsReponse, 0)
	for _, offering := range offerings {
		if offering.IsVisible {
			respOffering := &offeringsReponse{}
			respOffering.Offering = offering
			respOffering.OrganisationName = offering.Organisation.Name
			respOffering.OrganisationURL = offering.Organisation.Website
			respOfferings = append(respOfferings, respOffering)
		}
	}

	cigExchange.Respond(w, respOfferings)
}
