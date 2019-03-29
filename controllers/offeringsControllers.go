package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	models "cig-exchange-libs/models"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOffering handles GET organisations/{organisation_id}/offerings/{offering_id} endpoint
var GetOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetOffering)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// query offering from db
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// check if organisation id matches
	if offering.OrganisationID != organisationID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	cigExchange.Respond(w, offering)
}

// CreateOffering handles POST organisations/{organisation_id}/offerings endpoint
var CreateOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeCreateOffering)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.NewDecoder(r.Body).Decode(offering)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if offering.OrganisationID != organisationID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// insert offering into db
	apiError := offering.Create()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	cigExchange.Respond(w, offering)
}

// UpdateOffering handles PATCH organisations/{organisation_id}/offerings/{offering_id} endpoint
var UpdateOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeUpdateOffering)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		*apiErrorP = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	offering := &models.Offering{}
	// decode offering object from request body
	err = json.Unmarshal(bytes, offering)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	offeringMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &offeringMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// remove unknow fields from map
	filteredOfferingMap := cigExchange.FilterUnknownFields(&models.Offering{}, offeringMap)

	if len(offering.OrganisationID) == 0 {
		*apiErrorP = cigExchange.NewInvalidFieldError("organisation_id", "Required field 'organisation_id' missing")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if offering.OrganisationID != organisationID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	existingOffering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if existingOffering.OrganisationID != organisationID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// set the offering UUID
	offering.ID = offeringID
	filteredOfferingMap["id"] = offeringID

	// update offering
	apiError = offering.Update(filteredOfferingMap)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// return updated offering
	existingOffering, apiError = models.GetOffering(offeringID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, existingOffering)
}

// DeleteOffering handles DELETE organisations/{organisation_id}/offerings/{offering_id} endpoint
var DeleteOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeDeleteOffering)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// query offering from db first to validate the permissions
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if offering.OrganisationID != organisationID {
		*apiErrorP = cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// delete offering
	apiError = offering.Delete()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	w.WriteHeader(204)
}

// GetOfferings handles GET organisations/{organisation_id}/offerings endpoint
var GetOfferings = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetOfferings)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// query all offerings from db
	offerings, apiError := models.GetOrganisationOfferings(organisationID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	cigExchange.Respond(w, offerings)
}

// GetAllOfferings handles GET offerings endpoint
// does not perform JWT based organisation filtering
var GetAllOfferings = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeAllOfferings)
	defer cigExchange.PrintAPIError(apiErrorP)

	// query all offerings from db
	offerings, apiError := models.GetOfferings()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
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
