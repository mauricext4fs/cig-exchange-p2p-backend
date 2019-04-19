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

	apiErrorP, loggedInUserP := auth.PrepareActivityVariables()
	defer auth.CreateUserActivity(loggedInUserP, apiErrorP, models.ActivityTypeContactUs)
	defer cigExchange.PrintAPIError(apiErrorP)

	type contactUs struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	contactInfo := &contactUs{}
	// decode contact us info from request body
	err := json.NewDecoder(r.Body).Decode(contactInfo)
	if err != nil {
		*apiErrorP = cigExchange.NewRequestDecodingError(err)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
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
		*apiErrorP = cigExchange.NewRequiredFieldError(missingFieldNames)
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	mandrillClient := cigExchange.GetMandrill()

	recipients := []gochimp.Recipient{
		gochimp.Recipient{Email: os.Getenv("CONTACTUS_TARGET_EMAIL"), Name: "CIG Exchange team", Type: "to"},
	}

	message := gochimp.Message{
		Text:      fmt.Sprintf("Name:%s\nEmail:%s\n\nMessage:\n%s", contactInfo.Name, contactInfo.Email, contactInfo.Message),
		Subject:   "Contact Us message",
		FromEmail: os.Getenv("CONTACTUS_TARGET_EMAIL"),
		FromName:  "CIG Exchange contact us",
		To:        recipients,
	}

	resp, err := mandrillClient.MessageSend(message, false)

	// we only have 1 recepient
	if len(resp) == 1 {
		if resp[0].Status == "rejected" {
			*apiErrorP = &cigExchange.APIError{}
			(*apiErrorP).SetErrorType(cigExchange.ErrorTypeUnprocessableEntity)

			nesetedError := (*apiErrorP).NewNestedError(cigExchange.ReasonMandrillFailure, "Invalid request")
			nesetedError.OriginalError = fmt.Errorf("Invalid request. %v", resp[0].RejectedReason)
			cigExchange.RespondWithAPIError(w, *apiErrorP)
			return
		}
	} else {
		*apiErrorP = &cigExchange.APIError{}
		(*apiErrorP).SetErrorType(cigExchange.ErrorTypeInternalServer)

		nesetedError := (*apiErrorP).NewNestedError(cigExchange.ReasonMandrillFailure, "Unable to send email")
		nesetedError.OriginalError = fmt.Errorf("Unable to send email")
		cigExchange.RespondWithAPIError(w, *apiErrorP)
		return
	}

	w.WriteHeader(204)
}
