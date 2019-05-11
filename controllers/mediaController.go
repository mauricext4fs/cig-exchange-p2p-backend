package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"io/ioutil"
	"mime"
	"net/http"
	"path"

	"github.com/gorilla/mux"
)

// Size constants
const (
	MB = 1 << 20
)

// UserDataPath path to user data folder
const UserDataPath = "/user_data/"

// GetMedia handles GET media/{media_file} endpoint
var GetMedia = func(w http.ResponseWriter, r *http.Request) {

	// get request params
	mediaFile := mux.Vars(r)["media_file"]

	// make file url
	fileURL := path.Join(UserDataPath, mediaFile)

	http.ServeFile(w, r, fileURL)
}

// UploadMedia handles PUT organisations/{organisation_id}/offerings/{offering_id}/media/upload endpoint
var UploadMedia = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeUploadMedia)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]

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

	// query offering from db first to validate the permissions
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, 15*MB) // 15 Mb

	defer r.Body.Close()
	fileBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Failed to read request body", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	filetype := http.DetectContentType(fileBytes)

	// fill mediasize and type
	media := &models.MediaWithIndex{Media: &models.Media{}, Index: 100}
	media.FileSize = len(fileBytes)
	media.MimeType = filetype
	media.Type = models.MediaTypeDocument
	media.Index = 100

	exts, err := mime.ExtensionsByType(filetype)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Can't get file extension", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	if len(exts) > 0 {
		media.FileExtension = exts[0]
	}

	// insert offering into db
	apiError = models.CreateMediaForOffering(media, offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// generate media url
	media.URL = "/invest/api/media/" + media.ID + media.FileExtension

	// save new file url
	cigExchange.GetDB().Save(media)

	// save file to user_data folder
	err = ioutil.WriteFile(path.Join(UserDataPath, media.ID)+media.FileExtension, fileBytes, 0644)
	if err != nil {
		info.APIError = cigExchange.NewReadError("Failed to write request body to file", err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, media)
}

// GetOfferingMedia handles GET offerings/{offering_id}/media endpoint
var GetOfferingMedia = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeGetOfferingsMedia)
	defer cigExchange.PrintAPIError(info)

	// get request params
	offeringID := mux.Vars(r)["offering_id"]

	// query offering from db
	_, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// query all media for offering
	offeringMedias, apiError := models.GetMediaForOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	cigExchange.Respond(w, offeringMedias)
}

// UpdateOfferingMedia handles PATCH organisations/{organisation_id}/offerings/{offering_id}/media/{media_id} endpoint
var UpdateOfferingMedia = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeUpdateOfferingsMedia)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]
	mediaID := mux.Vars(r)["media_id"]

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

	// query offering from db
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// get exisitng media model
	media, apiError := models.GetMedia(mediaID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// get exisitng media model
	offeringMedia, apiError := models.GetOfferingMedia(offeringID, mediaID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// read media
	updateMedia := &models.Media{}
	original, filtered, apiError := cigExchange.ReadAndParseRequest(r.Body, updateMedia)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// check required field
	if len(updateMedia.Type) == 0 {
		info.APIError = cigExchange.NewInvalidFieldError("type", "Required field 'type' is missing")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	if updateMedia.Type != models.MediaTypeDocument && updateMedia.Type != models.MediaTypeImage {
		info.APIError = cigExchange.NewInvalidFieldError("type", "Required field 'type' can be only 'offering-image' or 'offering-document'")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// set ID
	filtered["id"] = mediaID

	// update media
	apiError = media.Update(filtered)
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

	// update media index
	apiError = offeringMedia.UpdateIndex(index)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	// prepare response
	mediaResponse := &models.MediaWithIndex{Media: media, Index: index}

	cigExchange.Respond(w, mediaResponse)
}

// DeleteOfferingMedia handles DELETE organisations/{organisation_id}/offerings/{offering_id}/media/{media_id} endpoint
var DeleteOfferingMedia = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeDeleteOfferingsMedia)
	defer cigExchange.PrintAPIError(info)

	// get request params
	organisationID := mux.Vars(r)["organisation_id"]
	offeringID := mux.Vars(r)["offering_id"]
	mediaID := mux.Vars(r)["media_id"]

	// load context user info
	loggedInUser, err := auth.GetContextValues(r)
	if err != nil {
		info.APIError = cigExchange.NewRoutingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	info.LoggedInUser = loggedInUser

	// query offering from db first to validate the permissions
	offering, apiError := models.GetOffering(offeringID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	if offering.OrganisationID != organisationID {
		info.APIError = cigExchange.NewAccessRightsError("Offering doesn't belong to organisation")
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
		_, apiError := models.GetOrgUserRole(loggedInUser.UserUUID, organisationID)
		if apiError != nil {
			// user don't belong to organisation
			info.APIError = apiError
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	}

	// delete offering
	apiError = models.DeleteOfferingMedia(mediaID)
	if apiError != nil {
		info.APIError = apiError
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}
	w.WriteHeader(204)
}
