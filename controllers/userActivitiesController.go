package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// GetUserActivities handles GET users/{user_id}/activities endpoints
var GetUserActivities = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetUserActivities)
	defer cigExchange.PrintAPIError(info)

	// get request params
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// check user id
	if len(userID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if userID != loggedInUser.UserUUID {
		info.APIError = cigExchange.NewAccessRightsError("No access rights for the user")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	userActs, apiError := models.GetActivitiesForUser(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, userActs)
}

// CreateUserActivity handles POST users/activities and users/{user_id}/activities endpoints
var CreateUserActivity = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer cigExchange.PrintAPIError(info)

	// check jwt
	loggedInUser, err := auth.GetContextValues(r)
	// ignore error for invest api call
	if err == nil {
		info.LoggedInUser = loggedInUser
	}

	infoMap := make(map[string]interface{})
	// decode organisation object from request body
	err = json.NewDecoder(r.Body).Decode(&infoMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		auth.CreateUserActivity(info, models.ActivityTypeCreateUserActivity)
		return
	}

	// insert user activity into db
	apiError := auth.CreateCustomUserActivity(info, infoMap)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		auth.CreateUserActivity(info, models.ActivityTypeCreateUserActivity)
		return
	}

	w.WriteHeader(204)
}
