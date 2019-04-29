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

// GetUser handles GET users/{user_id} endpoint
var GetUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetUser)
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

	// return user object
	existingUser, apiError := models.GetUser(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, existingUser)
}

// UpdateUser handles PATCH users/{user_id} endpoint
var UpdateUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r.RemoteAddr)
	defer auth.CreateUserActivity(info, models.ActivityTypeUpdateUser)
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

	// read request body
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	user := &models.User{}
	// decode user object from request body
	err = json.Unmarshal(bytes, user)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	userMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &userMap)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// remove unknow fields from map
	filteredUserMap := cigExchange.FilterUnknownFields(user, userMap)

	// set the user UUID
	user.ID = userID
	filteredUserMap["id"] = userID

	// update user
	apiError := user.Update(filteredUserMap)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// return updated user
	existingUser, apiError := models.GetUser(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, existingUser)
}
