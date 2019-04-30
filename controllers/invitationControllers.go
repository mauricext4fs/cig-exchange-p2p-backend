package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	models "cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type invitationRequest struct {
	Title            string `json:"title"`
	Name             string `json:"name"`
	LastName         string `json:"lastname"`
	Email            string `json:"email"`
	PhoneCountryCode string `json:"phone_country_code"`
	PhoneNumber      string `json:"phone_number"`
}

// SendInvitation handles POST organisations/{organisation_id}/invitations endpoint
var SendInvitation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeCreateInvitation)
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

	// check permissions
	if organisationID != loggedInUser.OrganisationUUID || len(organisationID) == 0 {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	userReq := &auth.UserRequest{}

	// decode user object from request body
	err = json.NewDecoder(r.Body).Decode(userReq)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check that organisation exists
	org, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check that user who invites exists
	inviter, apiError := models.GetUser(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	userReq.Platform = auth.PlatformP2P

	user := userReq.ConvertRequestToUser()

	// check if user exist
	invitedUser, apiError := models.GetUserByEmail(userReq.Email, true)
	if err != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	// invited user can be nill, which means it doesn't exist
	if invitedUser == nil {
		// create new user w/o reference key, we will create organisation link manually
		invitedUser, apiError = models.CreateUser(user, "")
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// check organisation link existance, we don't want double invites
	orgUserWhere := &models.OrganisationUser{
		UserID:         invitedUser.ID,
		OrganisationID: organisationID,
	}
	_, apiError = orgUserWhere.Find()
	if apiError == nil { // expecting error, no error means link exists
		apiError = &cigExchange.APIError{}
		apiError.SetErrorType(cigExchange.ErrorTypeUnprocessableEntity)
		apiError.NewNestedError(cigExchange.ReasonInvitationAlreadyExists, cigExchange.ReasonInvitationAlreadyExists)
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// create organisation link for the user
	orgUser := &models.OrganisationUser{
		UserID:           invitedUser.ID,
		OrganisationID:   organisationID,
		Status:           models.OrganisationUserStatusInvited,
		IsHome:           false,
		OrganisationRole: models.OrganisationRoleUser,
	}
	apiError = orgUser.Create()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// save the orgUser UUID into redis for accept workflow
	rediskey := cigExchange.RandomUUID()
	expiration := 30 * 24 * time.Hour
	redisCmd := cigExchange.GetRedis().Set(rediskey, orgUser.ID, expiration)
	if redisCmd.Err() != nil {
		info.APIError = cigExchange.NewRedisError("Set invitation accept code failure", redisCmd.Err())
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// send invitation email async
	go func() {
		// email parameters
		parameters := map[string]string{
			"ACCEPT_URL":           cigExchange.GetServerURL() + "/invest/en/#accept-invitation/" + rediskey,
			"INVITE_FIRST_NAME":    userReq.Name,
			"INVITER_NAME":         inviter.Name + " " + inviter.LastName,
			"INVITER_ORGANISATION": org.Name,
		}
		err = cigExchange.SendEmail(cigExchange.EmailTypeInvitation, userReq.Email, parameters)
		if err != nil {
			fmt.Println("InviteUser: email sending error:")
			fmt.Println(err.Error())
		}
	}()

	resp := make(map[string]string, 0)
	resp["uuid"] = invitedUser.ID
	// in "DEV" environment we return invitation accept code for testing purposes
	if cigExchange.IsDevEnv() {
		resp["code"] = rediskey
	}

	cigExchange.Respond(w, resp)
}

// GetInvitations handles GET organisations/{organisation_id}/invitations endpoint
var GetInvitations = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetInvitations)
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

	// check organisation id
	if len(organisationID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query invited users from db
	users, apiError := models.GetUsersForOrganisation(organisationID, true)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteInvitation handles DELETE organisations/{organisation_id}/invitations/{user_id} endpoint
var DeleteInvitation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteInvitation)
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

	// check organisation id
	if len(organisationID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check user id
	if len(userID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if organisationID != loggedInUser.OrganisationUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	searchOrgUser := &models.OrganisationUser{
		OrganisationID: organisationID,
		UserID:         userID,
	}
	// query user organisationUser from db
	orgUser, apiError := searchOrgUser.Find()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check user status
	if orgUser.Status != models.OrganisationUserStatusInvited {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "User already active")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check user organisation
	apiError = orgUser.Delete()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	w.WriteHeader(204)
}

// AcceptInvitation handles POST users/accept-invitation endpoint (no JWT)
var AcceptInvitation = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeAcceptInvitation)
	defer cigExchange.PrintAPIError(info)

	// get invitation accept key from post body
	acceptKeyMap := make(map[string]string)
	// decode map[string]string from request body
	err := json.NewDecoder(r.Body).Decode(&acceptKeyMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	acceptKey, ok := acceptKeyMap["invitation_id"]
	if !ok {
		info.APIError = cigExchange.NewRequiredFieldError([]string{"invitation_id"})
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	if len(acceptKey) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("invitation_id", "Invitation id cannot be empty")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	redisClient := cigExchange.GetRedis()
	redisCmd := redisClient.Get(acceptKey)
	if redisCmd.Err() != nil {
		info.APIError = cigExchange.NewRedisError("Unable to get invitation", redisCmd.Err())
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query user organisationUser from db
	orgUser, apiError := models.OrganisationUserByID(redisCmd.Val())
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if orgUser.Status != models.OrganisationUserStatusInvited {
		apiError = &cigExchange.APIError{}
		apiError.SetErrorType(cigExchange.ErrorTypeUnprocessableEntity)
		apiError.NewNestedError(cigExchange.ReasonInvitationAlreadyAccepted, cigExchange.ReasonInvitationAlreadyAccepted)
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query the user, validate his email if necessary
	user, apiError := models.GetUser(orgUser.UserID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	if user.Status == models.UserStatusUnverified {
		user.Status = models.UserStatusVerified
		apiError = user.Save()
		if apiError != nil {
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// mark invitation accepted
	orgUser.Status = models.OrganisationUserStatusActive
	apiError = orgUser.Update()
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// reply with JWT token
	tokenString, _, apiError := auth.GenerateJWTString(orgUser.UserID, orgUser.OrganisationID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	resp := &auth.JwtResponse{
		JWT: tokenString,
	}
	cigExchange.Respond(w, resp)
}
