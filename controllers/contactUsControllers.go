package controllers

import (
	cigExchange "cig-exchange-libs"
	"cig-exchange-libs/auth"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/mattbaird/gochimp"
)

// SendContactUsEmail handles POST api/contact_us endpoint
var SendContactUsEmail = func(w http.ResponseWriter, r *http.Request) {

	// create user activity record and print error with defer
	info := cigExchange.PrepareActivityInformation(r)
	defer auth.CreateUserActivity(info, models.ActivityTypeContactUs)
	defer cigExchange.PrintAPIError(info)

	type contactUs struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	contactInfo := &contactUs{}
	// decode contact us info from request body
	err := json.NewDecoder(r.Body).Decode(contactInfo)
	if err != nil {
		info.APIError = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	missingFieldNames := make([]string, 0)
	// check for empty request parameters
	if len(contactInfo.Name) == 0 {
		missingFieldNames = append(missingFieldNames, "name")
	}
	if len(contactInfo.Email) == 0 {
		missingFieldNames = append(missingFieldNames, "email")
	}
	if len(contactInfo.Message) == 0 {
		missingFieldNames = append(missingFieldNames, "message")
	}
	if len(missingFieldNames) > 0 {
		info.APIError = cigExchange.NewRequiredFieldError(missingFieldNames)
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	mandrillClient := cigExchange.GetMandrill()

	targetEmail := os.Getenv("CONTACTUS_TARGET_EMAIL")
	// blackhole contact us emails in dev env
	if cigExchange.IsDevEnv() {
		targetEmail = "blackhole+" + targetEmail
	}

	recipients := []gochimp.Recipient{
		gochimp.Recipient{Email: targetEmail, Name: "CIG Exchange team", Type: "to"},
	}

	message := gochimp.Message{
		Text:      fmt.Sprintf("Name:%s\nEmail:%s\n\nMessage:\n%s", contactInfo.Name, contactInfo.Email, contactInfo.Message),
		Subject:   "Contact Us message",
		FromEmail: targetEmail,
		FromName:  "CIG Exchange contact us",
		To:        recipients,
	}

	resp, err := mandrillClient.MessageSend(message, false)

	// we only have 1 recepient
	if len(resp) == 1 {
		if resp[0].Status == "rejected" {
			info.APIError = &cigExchange.APIError{}
			info.APIError.SetErrorType(cigExchange.ErrorTypeUnprocessableEntity)

			nesetedError := info.APIError.NewNestedError(cigExchange.ReasonMandrillFailure, "Invalid request")
			nesetedError.OriginalError = fmt.Errorf("Invalid request. %v", resp[0].RejectedReason)
			cigExchange.RespondWithAPIError(w, info.APIError)
			return
		}
	} else {
		info.APIError = &cigExchange.APIError{}
		info.APIError.SetErrorType(cigExchange.ErrorTypeInternalServer)

		nesetedError := info.APIError.NewNestedError(cigExchange.ReasonMandrillFailure, "Unable to send email")
		nesetedError.OriginalError = fmt.Errorf("Unable to send email")
		cigExchange.RespondWithAPIError(w, info.APIError)
		return
	}

	w.WriteHeader(204)
}
