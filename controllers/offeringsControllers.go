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
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetOffering)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query offering from db
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check if organisation id matches
	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	offeringMMap, apiError := cigExchange.PrepareResponseForMultilangModel(offering)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, offeringMMap)
}

// CreateOffering handles POST organisations/{organisation_id}/offerings endpoint
var CreateOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeCreateOffering)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	offering := &models.Offering{}

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	offeringMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &offeringMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// remove unknow fields from map
	filteredOfferingMap := cigExchange.FilterUnknownFields(offering, offeringMap)

	// convert multilang fields to jsonb
	cigExchange.ConvertRequestMapToJSONB(&filteredOfferingMap, offering)

	jsonBytes, err := json.Marshal(filteredOfferingMap)
	if err != nil {
		info.APIError = cigExchange.NewJSONEncodingError(cigExchange.MessageRequestJSONDecoding, err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, offering)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// insert offering into db
	apiError := offering.Create()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	offeringMMap, apiError := cigExchange.PrepareResponseForMultilangModel(offering)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, offeringMMap)
}

// UpdateOffering handles PATCH organisations/{organisation_id}/offerings/{offering_id} endpoint
var UpdateOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeUpdateOffering)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	offering := &models.Offering{}

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	offeringMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &offeringMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// remove unknow fields from map
	filteredOfferingMap := cigExchange.FilterUnknownFields(offering, offeringMap)

	// convert multilang fields to jsonb
	cigExchange.ConvertRequestMapToJSONB(&filteredOfferingMap, offering)

	jsonBytes, err := json.Marshal(filteredOfferingMap)
	if err != nil {
		info.APIError = cigExchange.NewJSONEncodingError(cigExchange.MessageRequestJSONDecoding, err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, offering)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if len(offering.OrganisationID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("organisation_id", "Required field 'organisation_id' missing")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	existingOffering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if existingOffering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// set the offering UUID
	offering.ID = offeringID
	filteredOfferingMap["id"] = offeringID

	// update offering
	apiError = offering.Update(filteredOfferingMap)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// return updated offering
	existingOffering, apiError = models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	offeringMMap, apiError := cigExchange.PrepareResponseForMultilangModel(existingOffering)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, offeringMMap)
}

// DeleteOffering handles DELETE organisations/{organisation_id}/offerings/{offering_id} endpoint
var DeleteOffering = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteOffering)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query offering from db first to validate the permissions
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// delete offering
	apiError = offering.Delete()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	w.WriteHeader(204)
}

// GetOfferings handles GET organisations/{organisation_id}/offerings endpoint
var GetOfferings = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetOfferings)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query all offerings from db
	offerings, apiError := models.GetOrganisationOfferings(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	offeringsAMap := make([]map[string]interface{}, 0)
	for _, offering := range offerings {
		offeringMMap, apiError := cigExchange.PrepareResponseForMultilangModel(offering)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
		offeringsAMap = append(offeringsAMap, offeringMMap)
	}

	cigExchange.Respond(w, offeringsAMap)
}

// GetAllOfferings handles GET offerings endpoint
// does not perform JWT based organisation filtering
var GetAllOfferings = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeAllOfferings)
	defer cigExchange.PrintAPIError(info)

	// query all offerings from db
	offerings, apiError := models.GetOfferings()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add organisation name to offerings structs
	offeringsAMap := make([]map[string]interface{}, 0)
	for _, offering := range offerings {
		if offering.IsVisible && offering.Organisation.Status != models.OrganisationStatusUnverified {
			// add multilang fields
			offeringMMap, apiError := cigExchange.PrepareResponseForMultilangModel(offering)
			if apiError != nil {
				info.APIError = apiError
				cigExchange.RespondWithAPIError(w, info.APIError)
				return
			}
			offeringMMap["organisation"] = offering.Organisation.Name
			offeringMMap["organisation_website"] = offering.Organisation.Website
			offeringsAMap = append(offeringsAMap, offeringMMap)
		}
	}

	cigExchange.Respond(w, offeringsAMap)
}
