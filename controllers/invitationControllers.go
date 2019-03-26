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
	Sex              string `json:"sex"`
	Name             string `json:"name"`
	LastName         string `json:"lastname"`
	Email            string `json:"email"`
	PhoneCountryCode string `json:"phone_country_code"`
	PhoneNumber      string `json:"phone_number"`
}

// SendInvitation handles POST organisations/{organisation_id}/invitations endpoint
var SendInvitation = func(w http.ResponseWriter, r *http.Request) {

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
	if organisationID != loggedInUser.OrganisationUUID || len(organisationID) == 0 {
		apiError := cigExchange.NewAccessRightsError("No access rights for the organisation")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	userReq := &auth.UserRequest{}

	// decode user object from request body
	err = json.NewDecoder(r.Body).Decode(userReq)
	if err != nil {
		apiError := cigExchange.NewJSONDecodingError(err)
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check that organisation exists
	org, apiError := models.GetOrganisation(organisationID)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	userReq.Platform = auth.PlatformP2P

	user := userReq.ConvertRequestToUser()

	// create user or invite existing user
	createdUser, apiError := models.CreateInvitedUser(user, org)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		if apiError.ShouldSilenceError() {
			w.WriteHeader(204)
		} else {
			cigExchange.RespondWithAPIError(w, apiError)
		}
		return
	}

	rediskey := cigExchange.GenerateRedisKey(createdUser.ID)
	expiration := 5 * time.Minute

	code := cigExchange.RandCode(6)
	redisCmd := cigExchange.GetRedis().Set(rediskey, code, expiration)
	if redisCmd.Err() != nil {
		apiError = cigExchange.NewRedisError("Set code failure", redisCmd.Err())
		fmt.Printf(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// send welcome email async
	go func() {
		err = cigExchange.SendEmail(cigExchange.EmailTypePinCode, userReq.Email, code)
		if err != nil {
			fmt.Println("InviteUser: email sending error:")
			fmt.Println(err.Error())
		}
	}()

	resp := make(map[string]string, 0)
	resp["uuid"] = createdUser.ID
	// in "DEV" environment we return the email signup code for testing purposes
	if cigExchange.IsDevEnv() {
		resp["code"] = code
	}

	cigExchange.Respond(w, resp)
}

// GetInvitations handles GET organisations/{organisation_id}/invitations endpoint
var GetInvitations = func(w http.ResponseWriter, r *http.Request) {

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

	// query invited users from db
	users, apiError := models.GetUsersForOrganisation(organisationID, true)
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	cigExchange.Respond(w, users)
}

// DeleteInvitation handles DELETE organisations/{organisation_id}/invitations/{user_id} endpoint
var DeleteInvitation = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	userID := mux.Vars(r)["user_id"]

	// check organisation id
	if len(organisationID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("organization_id", "OrganisationID is invalid")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check user id
	if len(userID) == 0 {
		apiError := cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
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

	searchOrgUser := &models.OrganisationUser{
		OrganisationID: organisationID,
		UserID:         userID,
	}
	// query user organisation from db
	orgUser, apiError := searchOrgUser.Find()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check user status
	if orgUser.Status != models.OrganisationUserStatusInvited {
		apiError = cigExchange.NewInvalidFieldError("user_id", "User already active")
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	// check user organisation
	apiError = orgUser.Delete()
	if apiError != nil {
		fmt.Println(apiError.ToString())
		cigExchange.RespondWithAPIError(w, apiError)
		return
	}

	w.WriteHeader(204)
}
