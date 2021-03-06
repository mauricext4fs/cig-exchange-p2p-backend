package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"net/http"

	"github.com/gorilla/mux"
)

// GetContacts handles GET users/{user_id}/contacts endpoint
var GetContacts = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetUserContacts)
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, loggedInUser.OrganisationUUID)
		if apiError != nil {
			// user don't belong to organisation
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}

		if loggedInUser.UserUUID != userID {
			// check target user
			_, apiError := models.GetOrgUserRole(userID, loggedInUser.OrganisationUUID)
			if apiError != nil {
				// user don't belong to organisation
				info.APIError = apiError
				cigExchange.RespondWithAPIError(w, info.APIError)
				return
			}
		}
	}

	// query offering from db
	contacts, apiError := models.GetContacts(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, contacts)
}

// CreateContact handles POST users/{user_id}/contacts endpoint
var CreateContact = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeCreateUserContact)
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

	// check admin
	userRole, apiError := models.GetUserRole(loggedInUser.UserUUID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// skip check for admin
	if userRole != models.UserRoleAdmin {
		if loggedInUser.UserUUID != userID {
			info.APIError = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// check that user exists
	_, apiError = models.GetUser(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// read contact
	contact := &models.Contact{}
	original, _, apiError := cigExchange.ReadAndParseRequest(r.Body, contact)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// parse index
	index, apiError := cigExchange.ParseIndex(original)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// can create only secondary level contacts for now
	contact.Level = models.ContactLevelSecondary

	// create contact and user_contact
	apiError = contact.Create(userID, index)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add index
	contactWithIndex := &models.ContactWithIndex{Contact: contact, Index: index}

	cigExchange.Respond(w, contactWithIndex)
}

// UpdateContact handles PATCH users/{user_id}/contacts/{contact_id} endpoint
var UpdateContact = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeUpdateUserContact)
	defer cigExchange.PrintAPIError(info)

	// get request params
	userID := mux.Vars(r)["user_id"]
	contactID := mux.Vars(r)["contact_id"]

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

	// check contact id
	if len(contactID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("contact_id", "ContactID is invalid")
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
		if loggedInUser.UserUUID != userID {
			info.APIError = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// read contact
	contact := &models.Contact{}
	original, filtered, apiError := cigExchange.ReadAndParseRequest(r.Body, contact)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// parse index
	index, apiError := cigExchange.ParseIndex(original)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// set contact id
	filtered["id"] = contactID
	contact.ID = contactID

	// can't change contact level for now
	delete(filtered, "level")

	// update coontact
	apiError = contact.Update(userID, filtered, index)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// add index
	contactWithIndex := &models.ContactWithIndex{Contact: contact, Index: index}

	cigExchange.Respond(w, contactWithIndex)
}

// DeleteContact handles DELETE users/{user_id}/contacts/{contact_id} endpoint
var DeleteContact = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteUserContact)
	defer cigExchange.PrintAPIError(info)

	// get request params
	userID := mux.Vars(r)["user_id"]
	contactID := mux.Vars(r)["contact_id"]

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

	// check contact id
	if len(contactID) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("contact_id", "ContactID is invalid")
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
		if loggedInUser.UserUUID != userID {
			info.APIError = cigExchange.NewInvalidFieldError("user_id", "UserID is invalid")
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// query contact from db
	contact, apiError := models.GetContact(contactID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if contact.Level == models.ContactLevelPrimary {
		info.APIError = cigExchange.NewInvalidFieldError("contact_id", "Can't delete primary contact")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// delete contact
	apiError = contact.Delete(userID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	w.WriteHeader(204)
}
