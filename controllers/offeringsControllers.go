package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	models "cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOffering handles GET organisations/{organisation_id}/offerings/{offering_id} endpoint
var GetOffering = func(w http.ResponseWriter, r *http.Request) {

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

	// query offering from db
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check if organisation id matches
	if offering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	cigExchange.Respond(w, offering)
}

// CreateOffering handles POST organisations/{organisation_id}/offerings endpoint
var CreateOffering = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

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

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if offering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// insert offering into db
	apiError := offering.Create()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	cigExchange.Respond(w, offering)
}

// UpdateOffering handles PATCH organisations/{organisation_id}/offerings/{offering_id} endpoint
var UpdateOffering = func(w http.ResponseWriter, r *http.Request) {

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

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiError := cigExchange.NewReadError("Failed to read request body", err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.Unmarshal(bytes, offering)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	offeringMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &offeringMap)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// remove unknow fields from map
	filteredOfferingMap := cigExchange.FilterUnknownFields(&models.Offering{}, offeringMap)

	if len(offering.OrganisationID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("organisation_id", "Required field 'organisation_id' missing")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if offering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	existingOffering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if existingOffering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// set the offering UUID
	offering.ID = offeringID
	filteredOfferingMap["id"] = offeringID

	// update offering
	apiError = offering.Update(filteredOfferingMap)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// return updated offering
	existingOffering, apiError = models.GetOffering(offeringID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, existingOffering)
}

// DeleteOffering handles DELETE organisations/{organisation_id}/offerings/{offering_id} endpoint
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
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if offering.OrganisationID != organisationID {
		apiError := cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// delete offering
	apiError = offering.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	w.WriteHeader(204)
}

// GetOfferings handles GET organisations/{organisation_id}/offerings endpoint
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

// GetAllOfferings handles GET offerings endpoint
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
