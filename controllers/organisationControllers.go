package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOrganisation handles GET organisations/{organisation_id} endpoint
var GetOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// check organisation id
	if len(organisationID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

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

	// query organisation from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, organisation)
}

// GetOrganisations handles GET organisations endpoint
var GetOrganisations = func(w http.ResponseWriter, r *http.Request) {

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	if len(loggedInUser.UserUUID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("user_id", "Invalid user id")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// query organisation from db
	organisations, apiError := models.GetOrganisations(loggedInUser.UserUUID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, organisations)
}

// CreateOrganisation handles POST organisations endpoint
var CreateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// check jwt
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	organisation := &models.Organisation{}
	// decode organisation object from request body
	err = json.NewDecoder(r.Body).Decode(organisation)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// insert organisation into db
	apiError := organisation.Create()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	orgUser := &models.OrganisationUser{
		UserID:           loggedInUser.UserUUID,
		OrganisationID:   organisation.ID,
		OrganisationRole: "admin",
		IsHome:           false,
	}

	// insert organisation user into db
	apiError = orgUser.Create()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, organisation)
}

// UpdateOrganisation handles PATCH organisations/{organisation_id} endpoint
var UpdateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// check organisation id
	if len(organisationID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

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

	organisation := &models.Organisation{}
	// decode organisation object from request body
	err = json.Unmarshal(bytes, organisation)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	organisationMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &organisationMap)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// remove unknow fields from map
	filteredOrganisationMap := cigExchange.FilterUnknownFields(&models.Organisation{}, organisationMap)

	// set the organisation UUID
	organisation.ID = organisationID
	filteredOrganisationMap["id"] = organisationID

	// update organisation
	apiError := organisation.Update(filteredOrganisationMap)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// return updated organisation
	existingOrganisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, existingOrganisation)
}

// DeleteOrganisation handles DELETE organisations/{organisation_id} endpoint
var DeleteOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]

	// check organisation id
	if len(organisationID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

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

	// query organisaion from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// delete organisation
	apiError = organisation.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// get all organisation user for organisation UUID
	orgUsers, apiError := models.GetOrganisationUsersForOrganisation(organisationID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// delete all organisation users
	for _, orgUser := range orgUsers {
		// TODO: hadnle JWT invalidation here
		apiError = orgUser.Delete()
		if apiError != nil {
			fmt.Println(apiError.ToString())
			cigExchange.RespondWithAPIError(w, apiError)
			return
		}
	}

	w.WriteHeader(204)
}

// GetOrganisationUsers handles GET organisations/{organisation_id}/users endpoint
var GetOrganisationUsers = func(w http.ResponseWriter, r *http.Request) {

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

	// check permissions
	if organisationID != loggedInUser.OrganisationUUID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// query users from db
	users, apiError := models.GetUsersForOrganisation(organisationID, false)
	if err != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteOrganisationUser handles DELETE organisations/{organisation_id}/users/{user_id} endpoint
var DeleteOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		apiError := cigExchange.NewRoutingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check permissions
	if organisationID != loggedInUser.OrganisationUUID {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// find user
	_, apiError := models.GetUser(userID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	// TODO: rewoke user JWT token

	// fill OrganizationUser with user id and organisation id
	searchOrgUser := models.OrganisationUser{
		UserID:         userID,
		OrganisationID: organisationID,
	}

	// find OrganizationUser
	orgUser, apiError := searchOrgUser.Find()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// delete OrganisationUser
	apiError = orgUser.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}
	w.WriteHeader(204)
}
