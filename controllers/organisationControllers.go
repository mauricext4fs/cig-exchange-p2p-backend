package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetOrganisation handles GET organisations/{organisation_id} endpoint
var GetOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetOrganisation)
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

	// query organisation from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	orgMap, apiError := cigExchange.PrepareResponseForMultilangModel(organisation)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, orgMap)
}

// GetOrganisations handles GET organisations endpoint
var GetOrganisations = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetOrganisations)
	defer cigExchange.PrintAPIError(info)

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	if len(loggedInUser.UserUUID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "Invalid user id")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// get user role
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// only admin user can create organisation
	if userRole == models.UserRoleAdmin {
		// query organisation from db
		organisations, apiError := models.GetAllOrganisations()
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}

		cigExchange.Respond(w, organisations)
		return
	}

	// query organisation from db
	organisations, apiError := models.GetOrganisations(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	orgsAMap := make([]map[string]interface{}, 0)
	for _, organisation := range organisations {
		orgMMap, apiError := cigExchange.PrepareResponseForMultilangModel(organisation)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
		orgsAMap = append(orgsAMap, orgMMap)
	}

	cigExchange.Respond(w, orgsAMap)
}

// CreateOrganisation handles POST organisations endpoint
var CreateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeCreateOrganisation)
	defer cigExchange.PrintAPIError(info)

	// check jwt
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// get user role
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// only admin user can create organisation
	if userRole != models.UserRoleAdmin {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	organisation := &models.Organisation{}
	organisationMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.NewDecoder(r.Body).Decode(&organisationMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// remove unknow fields from map
	filteredOrganisationMap := cigExchange.FilterUnknownFields(organisation, organisationMap)

	// convert multilang fields to jsonb
	cigExchange.ConvertRequestMapToJSONB(&filteredOrganisationMap, organisation)

	jsonBytes, err := json.Marshal(filteredOrganisationMap)
	if err != nil {
		info.APIError = cigExchange.NewJSONEncodingError(cigExchange.MessageRequestJSONDecoding, err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, organisation)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// insert organisation into db
	apiError = organisation.Create()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
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
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	orgMap, apiError := cigExchange.PrepareResponseForMultilangModel(organisation)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, orgMap)
}

// UpdateOrganisation handles PATCH organisations/{organisation_id} endpoint
var UpdateOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeUpdateOrganisation)
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
		orgUserRole, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}

		if orgUserRole != models.OrganisationRoleAdmin {
			info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	organisation := &models.Organisation{}
	organisationMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.NewDecoder(r.Body).Decode(&organisationMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// remove unknow fields from map
	filteredOrganisationMap := cigExchange.FilterUnknownFields(organisation, organisationMap)

	// convert multilang fields to jsonb
	cigExchange.ConvertRequestMapToJSONB(&filteredOrganisationMap, organisation)

	jsonBytes, err := json.Marshal(filteredOrganisationMap)
	if err != nil {
		info.APIError = cigExchange.NewJSONEncodingError(cigExchange.MessageRequestJSONDecoding, err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// decode offering object from request body
	err = json.Unmarshal(jsonBytes, organisation)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// set the organisation UUID
	organisation.ID = organisationID
	filteredOrganisationMap["id"] = organisationID

	// only admin can change organisation status
	if userRole != models.UserRoleAdmin {
		org, apiError := models.GetOrganisation(organisationID)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
		organisation.Status = org.Status
		filteredOrganisationMap["status"] = org.Status
	}

	// update organisation
	apiError = organisation.Update(filteredOrganisationMap)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// return updated organisation
	existingOrganisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add multilang fields
	orgMap, apiError := cigExchange.PrepareResponseForMultilangModel(existingOrganisation)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, orgMap)
}

// DeleteOrganisation handles DELETE organisations/{organisation_id} endpoint
var DeleteOrganisation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteOrganisation)
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

	// get user role and check user and organisation id
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// only admin user can delete organisation
	if userRole != models.UserRoleAdmin {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query organisaion from db
	organisation, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// delete organisation
	apiError = organisation.Delete()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// get all organisation user for organisation UUID
	orgUsers, apiError := models.GetOrganisationUsersForOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// delete all organisation users
	for _, orgUser := range orgUsers {
		// TODO: hadnle JWT invalidation here
		apiError = orgUser.Delete()
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	w.WriteHeader(204)
}

// GetOrganisationUsers handles GET organisations/{organisation_id}/users endpoint
var GetOrganisationUsers = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
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
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
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
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
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

// GetDashboardInfo handles GET organisations/{organisation_id}/dashboard endpoint
var GetDashboardInfo = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetDashboard)
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	dashboardInfo, apiError := models.GetOrganisationInfo(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, dashboardInfo)
}

// GetDashboardUsersInfo handles GET organisations/{organisation_id}/dashboard/users endpoint
var GetDashboardUsersInfo = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetDashboardUsers)
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	dashboardInfo, apiError := models.GetOrganisationUsersInfo(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, dashboardInfo)
}

// GetDashboardOfferingsBreakdown handles GET organisations/{organisation_id}/dashboard/offerings endpoint
var GetDashboardOfferingsBreakdown = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetDashboardBreakdown)
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	dashboardInfo, apiError := models.GetOfferingsTypeBreakdown(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, dashboardInfo)
}

// GetDashboardOfferingsClicks handles GET organisations/{organisation_id}/dashboard/clicks endpoint
var GetDashboardOfferingsClicks = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetDashboardClick)
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	dashboardInfo, apiError := models.GetOfferingsClicks(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, dashboardInfo)
}
