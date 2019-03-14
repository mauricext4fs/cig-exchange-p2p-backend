package controllers

import (
	"cig-exchange-libs"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mattbaird/gochimp"
)

// SendContactUsEmail handles POST api/contact_us endpoint
var SendContactUsEmail = func(w http.ResponseWriter, r *http.Request) {

	type contactUs struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Message string `json:"message"`
	}
	contactInfo := &contactUs{}
	// decode contact us info from request body
	err := json.NewDecoder(r.Body).Decode(contactInfo)
	if err != nil {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request"))
		return
	}

	// check for empty request parameters
	if len(contactInfo.Name) == 0 || len(contactInfo.Email) == 0 || len(contactInfo.Message) == 0 {
		cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request. %v", contactInfo))
		return
	}

	mandrillClient := cigExchange.GetMandrill()

	recipients := []gochimp.Recipient{
		gochimp.Recipient{Email: "info@cig-exchange.ch", Name: "CIG Exchange team", Type: "to"},
	}

	message := gochimp.Message{
		Text:      fmt.Sprintf("Name:%s\nEmail:%s\n\nMessage:\n%s", contactInfo.Name, contactInfo.Email, contactInfo.Message),
		Subject:   "Contact Us message",
		FromEmail: "info@cig-exchange.ch",
		FromName:  "CIG Exchange contact us",
		To:        recipients,
	}

	resp, err := mandrillClient.MessageSend(message, false)

	// we only have 1 recepient
	if len(resp) == 1 {
		if resp[0].Status == "rejected" {
			cigExchange.RespondWithError(w, 422, fmt.Errorf("Invalid request. %v", resp[0].RejectedReason))
			return
		}
	} else {
		cigExchange.RespondWithError(w, 500, fmt.Errorf("Unable to send email"))
		return
	}

	w.WriteHeader(204)
}
