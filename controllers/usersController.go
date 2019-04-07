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
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeGetUser)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	// check user id
	if len(userID) == 0 {
		*apiErrorP = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if userID != loggedInUser.UserUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the user")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// return user object
	existingUser, apiError := models.GetUser(userID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, existingUser)
}

// UpdateUser handles PATCH users/{user_id} endpoint
var UpdateUser = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeUpdateUser)
	defer cigExchange.PrintAPIError(apiErrorP)

	// get request params
	userID := mux.Vars(r)["user_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		*apiErrorP = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}
	*loggedInUserP = loggedInUser

	// check user id
	if len(userID) == 0 {
		*apiErrorP = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	if userID != loggedInUser.UserUUID {
		*apiErrorP = cigExchange.NewAccessRightsError("No access rights for the user")
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

	user := &models.User{}
	// decode user object from request body
	err = json.Unmarshal(bytes, user)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	userMap := make(map[string]interface{})
	// decode map[string]interface from request body
	err = json.Unmarshal(bytes, &userMap)
	if err != nil {
		*apiErrorP = cigExchange.NewJSONDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// remove unknow fields from map
	filteredUserMap := cigExchange.FilterUnknownFields(&models.User{}, []string{}, userMap)

	// set the user UUID
	user.ID = userID
	filteredUserMap["id"] = userID

	// update user
	apiError := user.Update(filteredUserMap)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	// return updated user
	existingUser, apiError := models.GetUser(userID)
	if apiError != nil {
		*apiErrorP = apiError
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	cigExchange.Respond(w, existingUser)
}
