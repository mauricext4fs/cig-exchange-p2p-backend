package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	models "cig-exchange-libs/models"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm/dialects/postgres"
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

	offeringMMap, apiError := prepareOfferingResponse(offering)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, offeringMMap)
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

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		*apiErrorP = cigExchange.NewReadError("Failed to read request body", err)
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
	filteredOfferingMap := cigExchange.FilterUnknownFields(&models.Offering{}, offering.GetMultilangFields(), offeringMap)

	// get jsonb fields
	names := offering.GetMultilangFields()
	convertMapToJSONB(&filteredOfferingMap, names)

	jsonBytes, err := json.Marshal(filteredOfferingMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONEncodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, offering)
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

	offeringMMap, apiError := prepareOfferingResponse(offering)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, offeringMMap)
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

	offering := &models.Offering{}

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

	offeringMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &offeringMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// remove unknow fields from map
	filteredOfferingMap := cigExchange.FilterUnknownFields(&models.Offering{}, offering.GetMultilangFields(), offeringMap)

	// get jsonb fields
	names := offering.GetMultilangFields()
	convertMapToJSONB(&filteredOfferingMap, names)

	jsonBytes, err := json.Marshal(filteredOfferingMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONEncodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, offering)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

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

	offeringMMap, apiError := prepareOfferingResponse(existingOffering)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, offeringMMap)
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

	offeringsAMap := make([]map[string]interface{}, 0)
	for _, offering := range offerings {
		offeringMMap, apiError := prepareOfferingResponse(offering)
		if apiError != nil {
			*apiErrorP = apiError
			cigExchange.RespondWithAPIError(w, *apiErrorP)
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

	// add organisation name to offerings structs
	offeringsAMap := make([]map[string]interface{}, 0)
	for _, offering := range offerings {
		if offering.IsVisible && offering.Organisation.Status != models.OrganisationStatusUnverified {
			offeringMMap, apiError := prepareOfferingResponse(offering)
			if apiError != nil {
				*apiErrorP = apiError
				cigExchange.RespondWithAPIError(w, *apiErrorP)
				return
			}
			offeringMMap["organisation"] = offering.Organisation.Name
			offeringMMap["organisation_website"] = offering.Organisation.Website
			offeringsAMap = append(offeringsAMap, offeringMMap)
		}
	}

	cigExchange.Respond(w, offeringsAMap)
}

func convertMapToJSONB(offeringMap *map[string]interface{}, fields []string) *cigExchange.APIError {

	localMap := *offeringMap

	for _, name := range fields {
		val, ok := localMap[name]
		if !ok {
			continue
		}
		switch v := val.(type) {
		case string:
			strVal := `{"en":"` + v + `"}`
			metadata := json.RawMessage(strVal)
			localMap[name] = postgres.Jsonb{RawMessage: metadata}
		case int32, int64:
			return cigExchange.NewInvalidFieldError(name, "Field '"+name+"' has invalid type")
		default:
			mapB, err := json.Marshal(v)
			if err != nil {
				return cigExchange.NewJSONEncodingError(err)
			}
			metadata := json.RawMessage(mapB)
			localMap[name] = postgres.Jsonb{RawMessage: metadata}
		}
	}
	return nil
}

func prepareOfferingResponse(offering *models.Offering) (map[string]interface{}, *cigExchange.APIError) {

	offeringMap := make(map[string]interface{})
	// marshal to json
	offeringBytes, err := json.Marshal(offering)
	if err != nil {
		return offeringMap, cigExchange.NewJSONEncodingError(err)
	}

	// fill map
	err = json.Unmarshal(offeringBytes, &offeringMap)
	if err != nil {
		return offeringMap, cigExchange.NewJSONDecodingError(err)
	}

	// handle multilanguage text
	for _, name := range offering.GetMultilangFields() {
		val, ok := offeringMap[name]
		if !ok {
			continue
		}
		// move jsonb to name_map field
		offeringMap[name+"_map"] = val
		// search for 'en' in jsonb
		offeringMap[name] = ""
		if val != nil {
			mapLang, ok := val.(map[string]interface{})
			if ok {
				if v, ok := mapLang["en"]; ok {
					offeringMap[name] = v
				} else if v, ok := mapLang["fr"]; ok {
					offeringMap[name] = v
				} else if v, ok := mapLang["fr"]; ok {
					offeringMap[name] = v
				} else if v, ok := mapLang["fr"]; ok {
					offeringMap[name] = v
				}
			}
		}
	}

	return offeringMap, nil
}
