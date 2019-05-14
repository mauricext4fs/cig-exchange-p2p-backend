package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOrganisationUsers handles GET organisations/{organisation_id}/users endpoint
var GetOrganisationUsers = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetUsers)
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

	// check admin
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// query users from db
	users, apiError := models.GetUsersForOrganisation(organisationID, false)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteOrganisationUser handles DELETE organisations/{organisation_id}/users/{user_id} endpoint
var DeleteOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteUser)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// fill OrganizationUser with user id and organisation id
	searchOrgUser := models.OrganisationUser{
		UserID:         userID,
		OrganisationID: organisationID,
	}

	// find OrganizationUser
	orgUserDelete, apiError := searchOrgUser.Find()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check admin
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		// check organisation role
		orgUserRole, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}

		if orgUserRole != models.OrganisationRoleAdmin {
			info.APIError = cigExchange.NewAccessRightsError("Only admin user can delete users from organisation")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}

		// checks before delete admin user
		if orgUserDelete.OrganisationRole == models.OrganisationRoleAdmin {

			if orgUserDelete.UserID == loggedInUser.UserUUID {
				info.APIError = cigExchange.NewAccessRightsError("Admin user can't remove himself from organisation")
				cigExchange.RespondWithAPIError(w, info.APIError)
				return
			}
		}
	}

	// delete OrganisationUser
	apiError = orgUserDelete.Delete()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	w.WriteHeader(204)
}

// AddOrganisationUser handles POST organisations/{organisation_id}/users/{user_id} endpoint
var AddOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeAddUser)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// check admin
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		info.APIError = cigExchange.NewAccessRightsError("Only admin user can directly add users to organisation. Organisation admin must use '/organisation/{}/invitations' api calls")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// fill OrganizationUser with user id and organisation id
	searchOrgUser := models.OrganisationUser{
		UserID:         userID,
		OrganisationID: organisationID,
	}

	// find OrganizationUser
	_, apiErrorTemp := searchOrgUser.Find()
	if apiErrorTemp == nil {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "User already belongs to organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	orgUser := models.OrganisationUser{
		UserID:           userID,
		OrganisationID:   organisationID,
		IsHome:           false,
		OrganisationRole: models.OrganisationRoleUser,
		Status:           models.OrganisationStatusUnverified,
	}

	// create OrganisationUser
	apiError = orgUser.Create()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	w.WriteHeader(204)
}

type changeOrganisationRequest struct {
	IsAdmin bool `json:"is_admin"`
}

// ChangeOrganisationUser handles PATCH organisations/{organisation_id}/users/{user_id} endpoint
var ChangeOrganisationUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypePatchUser)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// read request body
	changeOrgRequest := &changeOrganisationRequest{}

	err = json.NewDecoder(r.Body).Decode(&changeOrgRequest)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// find all OrganizationUser
	orgUsers, apiError := models.GetOrganisationUsersForOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// create checks variables
	adminCheck := false
	adminCount := 0
	var targetOrgUser *models.OrganisationUser
	for _, orgUser := range orgUsers {
		// count all admin users
		if orgUser.OrganisationRole == models.OrganisationRoleAdmin {
			adminCount++
			if orgUser.UserID == loggedInUser.UserUUID {
				// logged in user is admin
				adminCheck = true
			}
		}
		if orgUser.UserID == userID {
			// user belogs to organisation
			targetOrgUser = orgUser
		}
	}

	// check logged in user
	if !adminCheck {
		info.APIError = cigExchange.NewAccessForbiddenError("Only organisation admin can change organisation roles")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check target user
	if targetOrgUser == nil {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "User doesn't belong to organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// set admin
	if changeOrgRequest.IsAdmin {
		// skip if current role is admin
		if targetOrgUser.OrganisationRole == models.OrganisationRoleAdmin {
			w.WriteHeader(204)
			return
		}
		targetOrgUser.OrganisationRole = models.OrganisationRoleAdmin
	} else {
		// unset admin
		// skip if current role is not admin
		if targetOrgUser.OrganisationRole != models.OrganisationRoleAdmin {
			w.WriteHeader(204)
			return
		}
		targetOrgUser.OrganisationRole = models.OrganisationRoleUser
		// check for last admin
		if adminCount < 2 {
			info.APIError = cigExchange.NewAccessForbiddenError("Can't unset last organisation admin.")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// save new role
	apiError = targetOrgUser.Update()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	w.WriteHeader(204)
}
