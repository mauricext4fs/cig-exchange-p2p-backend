package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOrganisation handles GET organisations/{organisation_id} endpoint
var GetOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetOrganisation)
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

	// check admin
	userRole, apiError := auth.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		_, apiError := auth.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			*apiErrorP = apiError
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}
	}

	// query organisation from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, organisation)
}

// GetOrganisations handles GET organisations endpoint
var GetOrganisations = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetOrganisation)
	defer cigExchange.PrintAPIError(apiErrorP)

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	if len(loggedInUser.UserUUID) == 0 {
		*apiErrorP = cigExchange.NewInvalidFieldError("user_id", "Invalid user id")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// query organisation from db
	organisations, apiError := models.GetOrganisations(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, organisations)
}

// CreateOrganisation handles POST organisations endpoint
var CreateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeCreateOrganisation)
	defer cigExchange.PrintAPIError(apiErrorP)

	// check jwt
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	organisation := &models.Organisation{}
	// decode organisation object from request body
	err = json.NewDecoder(r.Body).Decode(organisation)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// insert organisation into db
	apiError := organisation.Create()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	orgUser := &models.OrganisationUser{
		UserID:           loggedInUser.UserUUID,
		OrganisationID:   organisation.ID,
		OrganisationRole: models.OrganisationRoleAdmin,
		IsHome:           false,
	}

	// insert organisation user into db
	apiError = orgUser.Create()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, organisation)
}

// UpdateOrganisation handles PATCH organisations/{organisation_id} endpoint
var UpdateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeUpdateOrganisation)
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

	// check admin
	userRole, apiError := auth.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		orgUserRole, apiError := auth.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			*apiErrorP = apiError
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}

		if orgUserRole != models.OrganisationRoleAdmin {
			*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}
	}

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		*apiErrorP = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	organisation := &models.Organisation{}
	// decode organisation object from request body
	err = json.Unmarshal(bytes, organisation)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	organisationMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &organisationMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// remove unknow fields from map
	filteredOrganisationMap := cigExchange.FilterUnknownFields(&models.Organisation{}, organisationMap)

	// set the organisation UUID
	organisation.ID = organisationID
	filteredOrganisationMap["id"] = organisationID

	// update organisation
	apiError = organisation.Update(filteredOrganisationMap)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// return updated organisation
	existingOrganisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, existingOrganisation)
}

// DeleteOrganisation handles DELETE organisations/{organisation_id} endpoint
var DeleteOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeDeleteOrganisation)
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

	// get user role and check user and organisation id
	userRole, apiError := auth.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// only admin user can delete organisation
	if userRole != models.UserRoleAdmin {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// query organisaion from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// delete organisation
	apiError = organisation.Delete()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// get all organisation user for organisation UUID
	orgUsers, apiError := models.GetOrganisationUsersForOrganisation(organisationID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// delete all organisation users
	for _, orgUser := range orgUsers {
		// TODO: hadnle JWT invalidation here
		*apiErrorP = orgUser.Delete()
		if *apiErrorP != nil {
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}
	}

	w.WriteHeader(204)
}

// GetOrganisationUsers handles GET organisations/{organisation_id}/users endpoint
var GetOrganisationUsers = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetUsers)
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

	// check admin
	userRole, apiError := auth.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		_, apiError := auth.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			*apiErrorP = apiError
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}
	}

	// query users from db
	users, apiError := models.GetUsersForOrganisation(organisationID, false)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteOrganisationUser handles DELETE organisations/{organisation_id}/users/{user_id} endpoint
var DeleteOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeDeleteUser)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	// fill OrganizationUser with user id and organisation id
	searchOrgUser := models.OrganisationUser{
		UserID:         userID,
		OrganisationID: organisationID,
	}

	// find OrganizationUser
	orgUserDelete, apiError := searchOrgUser.Find()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// check admin
	userRole, apiError := auth.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		orgUserRole, apiError := auth.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			*apiErrorP = apiError
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}

		if orgUserRole != models.OrganisationRoleAdmin {
			*apiErrorP = cigExchange.NewAccessRightsError("Only admin user can delete users from organisation")
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}

		// checks before delete admin user
		if orgUserDelete.OrganisationRole == models.OrganisationRoleAdmin {

			if orgUserDelete.UserID == loggedInUser.UserUUID {
				*apiErrorP = cigExchange.NewAccessRightsError("Admin user can't remove himself from organisation")
				cigExchange.RespondWithAPIError(w, *apiErrorP)
				return
			}
		}
	}

	// delete OrganisationUser
	apiError = orgUserDelete.Delete()
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	w.WriteHeader(204)
}
