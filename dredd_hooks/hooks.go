package main

import (
	"bytes"
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"

	cigExchange "cig-exchange-libs"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

const dredd = "dredd"
const dredd2 = "dredd2"
const dredd3 = "dredd3"
const dredd4 = "dredd4"
const dreddP2P = "dredd_p2p"

func main() {

	h := hooks.NewHooks()
	server := hooks.NewServer(hooks.NewHooksRunner(h))

	redisClient := cigExchange.GetRedis()
	dbClient := cigExchange.GetDB()

	// save created user UUID, JWT and organization UUID
	userUUID := ""
	userJWT := ""
	orgUUID := ""
	mediaUUID := ""

	// save created record UUID here
	createdUUID := ""
	invitationUUID := ""
	invitedUserUUID := ""
	invitationCode := ""
	contactUUID := ""

	// prepare the database:
	// 1. delete 'dredd' users if it exists (first name  = 'dredd', 'dredd2', 'dredd3', 'dredd4')
	// 2. delete 'dredd' organisations if it exists (reference key = 'dredd', 'dredd2', 'dredd3')
	// 3. create 'dredd' organisation (user will be registered with it)
	// 4. create some offerings belonging to 'dredd' organization
	// 5. verify that created offerings are present in 'invest/offerings' api call
	// 6. create dredd4 user
	// 7. add 'dredd4' user to organisation 'dredd'

	// delete 'dredd' users if it exists (first name  = 'dredd', 'dredd2', 'dredd3')
	dreddUsers := [4]string{dredd, dredd2, dredd3, dredd4}
	for _, name := range dreddUsers {
		usersDelete := make([]models.User, 0)
		err := dbClient.Where(&models.User{Name: name}).Find(&usersDelete).Error
		if err == nil {
			for _, u := range usersDelete {
				orgUsersDelete := make([]models.OrganisationUser, 0)
				err = dbClient.Where(&models.OrganisationUser{UserID: u.ID}).Find(&orgUsersDelete).Error
				if err == nil {
					for _, ou := range orgUsersDelete {
						dbClient.Delete(&ou)
					}
				}

				userContactDelete := make([]models.UserContact, 0)
				err = dbClient.Where(&models.UserContact{UserID: u.ID}).Find(&userContactDelete).Error
				if err == nil {
					for _, uc := range userContactDelete {
						dbClient.Delete(&models.Contact{ID: uc.ContactID})

						dbClient.Delete(&uc)
					}
					dbClient.Delete(&u)
				}

				dbClient.Delete(&u)
			}
		}
	}

	// delete 'dredd' organisations if it exists (reference key = 'dredd', 'dredd2', 'dredd3')
	for _, orgReference := range dreddUsers {
		orgsDelete := make([]models.Organisation, 0)
		err := dbClient.Where(&models.Organisation{ReferenceKey: orgReference}).Find(&orgsDelete).Error
		if err == nil {
			for _, o := range orgsDelete {
				dbClient.Delete(&o)
			}
		}
	}

	metadata := json.RawMessage(`{"en":"url","fr":"url","it":"url","de":"url"}`)
	titleMetadata := json.RawMessage(`{"en":"` + dredd + `"}`)

	// create 'dredd' organisation
	org := models.Organisation{
		Name:                      dredd,
		ReferenceKey:              dredd,
		Status:                    models.OrganisationStatusVerified,
		OfferingRatingDescription: postgres.Jsonb{RawMessage: titleMetadata},
	}
	err := dbClient.Create(&org).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create organisation:")
		fmt.Println(err.Error())
	}
	orgUUID = org.ID

	// create some offerings belonging to 'dredd' organization
	amount := float64(100)
	taken := float64(50)
	offering := models.Offering{
		Title:              postgres.Jsonb{RawMessage: titleMetadata},
		OrganisationID:     orgUUID,
		Type:               make(pq.StringArray, 0),
		IsVisible:          true,
		OfferingDirectURL:  postgres.Jsonb{RawMessage: metadata},
		Origin:             "origin",
		Amount:             &amount,
		Remaining:          float64(50),
		AmountAlreadyTaken: &taken,
	}

	err = dbClient.Create(&offering).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create offering:")
		fmt.Println(err.Error())
	}
	offeringID := offering.ID

	// add 'dredd4' user
	dredd4 := &models.User{
		Title:      "Mr",
		Role:       models.UserRoleUser,
		Name:       "dredd4",
		LastName:   "dredd4",
		LoginEmail: &models.Contact{Type: models.ContactTypeEmail, Level: models.ContactLevelPrimary, Value1: "blackhole+dev+test+dredd@cig-exchange.ch"},
		LoginPhone: &models.Contact{Type: models.ContactTypePhone, Level: models.ContactLevelSecondary, Value1: "+1", Value2: "333333333"},
		Status:     models.UserStatusVerified,
	}
	err = dbClient.Create(&dredd4).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create dredd4:")
		fmt.Println(err.Error())
	}

	// add 'dredd4' user to organisation 'dredd'
	orgUserDD := &models.OrganisationUser{
		UserID:           dredd4.ID,
		OrganisationID:   orgUUID,
		Status:           models.OrganisationUserStatusActive,
		IsHome:           true,
		OrganisationRole: models.OrganisationRoleUser,
	}
	err = dbClient.Create(&orgUserDD).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create organisation user:")
		fmt.Println(err.Error())
	}

	// inject JWT auth into api calls (if JWT is present)
	h.BeforeEach(func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}

		if len(userJWT) == 0 {
			return
		}

		t.Request.Headers["Authorization"] = fmt.Sprintf("Bearer %v", userJWT)
	})

	h.After("Trading/Offerings > invest/api/offerings > Offerings", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		// verify that created offerings are present in 'invest/offerings' api call
		found := bytes.Contains([]byte(t.Real.Body), []byte(dredd))
		if found {
			return
		}

		t.Fail = "Pre-created offering is missing"
	})

	h.Before("Trading/Users > invest/api/users/signup > Signup", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		setBodyValue(&t.Request.Body, "reference_key", dredd)
		setBodyValue(&t.Request.Body, "name", dredd)
	})

	h.After("Trading/Users > invest/api/users/signin > Signin", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		userUUID = getBodyValue(&t.Real.Body, "uuid")
		if len(userUUID) == 0 {
			t.Fail = "Unable to save user UUID"
			return
		}
	})

	h.Before("Trading/Users > invest/api/users/send_otp > Send OTP", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(userUUID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		setBodyValue(&t.Request.Body, "uuid", userUUID)
	})

	h.Before("Trading/Users > invest/api/users/signup/{user}/webauthn > Web Authn Signup", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		t.Skip = true
	})

	h.Before("Trading/Users > invest/api/users/signin/{user}/webauthn > Web Authn Signin", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		t.Skip = true
	})

	h.BeforeValidation("Trading/Users > invest/api/users/send_otp > Send OTP", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		if t.Real.StatusCode != 200 {
			return
		}

		// ignore dev env status code return
		t.Real.StatusCode = 204
		t.Real.Body = ""
	})

	h.Before("Trading/Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(userUUID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		rediskey := fmt.Sprintf("%s_signup_key", userUUID)
		redisCmd := redisClient.Get(rediskey)
		if redisCmd.Err() != nil {
			t.Fail = fmt.Sprintf("Redis error: %v", redisCmd.Err().Error())
			return
		}

		setBodyValue(&t.Request.Body, "uuid", userUUID)
		setBodyValue(&t.Request.Body, "code", redisCmd.Val())
		err := dbClient.Model(&models.User{}).Updates(&models.User{ID: userUUID, Role: models.UserRoleAdmin}).Error
		if err != nil {
			fmt.Println("ERROR: updatingDatabase: create admin:")
			fmt.Println(err.Error())
		}
	})

	h.After("Trading/Users > invest/api/users/verify_otp > Verify OTP", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		userJWT = getBodyValue(&t.Real.Body, "jwt")
		if len(userJWT) == 0 {
			t.Fail = "Unable to save user JWT"
			return
		}
	})

	h.Before("Trading/Organisations > invest/api/organisations/signup > Create organisation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}

		setBodyValue(&t.Request.Body, "name", dredd3)
		setBodyValue(&t.Request.Body, "organisation_name", dredd3)
		setBodyValue(&t.Request.Body, "reference_key", dredd3)
	})

	h.After("Trading/Organisations > invest/api/organisations/signup > Create organisation", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		invitedUserUUID = getBodyValue(&t.Real.Body, "uuid")
		if len(invitedUserUUID) == 0 {
			t.Fail = "Unable to save user UUID"
			return
		}
	})

	h.Before("P2P/Users > p2p/api/users/switch/{organisation} > Switch Organisation", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/users/switch/" + orgUUID
		t.FullPath = "/p2p/api/users/switch/" + orgUUID
	})

	h.After("P2P/Users > p2p/api/users/switch/{organisation} > Switch Organisation", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		userJWT = getBodyValue(&t.Real.Body, "jwt")
		if len(userJWT) == 0 {
			t.Fail = "Unable to save user JWT"
			return
		}
	})

	h.Before("P2P/Users > p2p/api/users/{user} > Retrieve user", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID
		t.FullPath = "/p2p/api/users/" + userUUID
	})

	h.Before("P2P/Users > p2p/api/users/{user} > Update user", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		setBodyValue(&t.Request.Body, "name", dredd)
		t.Request.URI = "/p2p/api/users/" + userUUID
		t.FullPath = "/p2p/api/users/" + userUUID
	})

	h.Before("P2P/Users > p2p/api/users/{user}/contacts > Retrieve user contacts", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID + "/contacts"
		t.FullPath = "/p2p/api/users/" + userUUID + "/contacts"
	})

	h.Before("P2P/Users > p2p/api/users/{user}/contacts > Create user contact", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID + "/contacts"
		t.FullPath = "/p2p/api/users/" + userUUID + "/contacts"
	})

	h.After("P2P/Users > p2p/api/users/{user}/contacts > Create user contact", func(t *trans.Transaction) {

		if t.Real == nil {
			return
		}

		contactUUID = getBodyValue(&t.Real.Body, "id")
	})

	h.Before("P2P/Users > p2p/api/users/{user}/contacts/{contact} > Update user contact", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(contactUUID) == 0 {
			t.Fail = "Contact UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID + "/contacts/" + contactUUID
		t.FullPath = "/p2p/api/users/" + userUUID + "/contacts/" + contactUUID
	})

	h.Before("P2P/Users > p2p/api/users/{user}/contacts/{contact} > Delete user contact", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(contactUUID) == 0 {
			t.Fail = "Contact UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID + "/contacts/" + contactUUID
		t.FullPath = "/p2p/api/users/" + userUUID + "/contacts/" + contactUUID
	})

	h.Before("P2P/Users > p2p/api/users/{user}/activities > Retrieve user activities", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/users/" + userUUID + "/activities"
		t.FullPath = "/p2p/api/users/" + userUUID + "/activities"
	})

	h.Before("P2P/Dashboard > p2p/api/organisations/{organisation}/dashboard > Get organisation info", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/dashboard"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/dashboard"
	})

	h.Before("P2P/Dashboard > p2p/api/organisations/{organisation}/dashboard/users > Get organisation users info", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/dashboard/users"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/dashboard/users"
	})

	h.Before("P2P/Dashboard > p2p/api/organisations/{organisation}/dashboard/offerings > Get organisation offerings breakdown", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/dashboard/offerings"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/dashboard/offerings"
	})

	h.Before("P2P/Dashboard > p2p/api/organisations/{organisation}/dashboard/clicks > Get organisation offerings clicks", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/dashboard/clicks"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/dashboard/clicks"
	})

	h.Before("P2P/Organisations > p2p/api/organisations > Create organisation", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		setBodyValue(&t.Request.Body, "reference_key", dredd2)
		setBodyValue(&t.Request.Body, "name", dredd2)
	})

	h.Before("P2P/Organisations > p2p/api/organisations/{organisation} > Retrieve organisation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Created organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID
	})

	h.Before("P2P/Organisations > p2p/api/organisations/{organisation} > Update organisation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Created organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID

		setBodyValue(&t.Request.Body, "name", dredd+cigExchange.RandCode(4))
	})

	h.Before("P2P/Organisations > p2p/api/organisations/{organisation} > Delete organisation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Created organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID
	})

	// update URI everywhere to point to a created record
	h.Before("P2P/Offerings > p2p/api/organisations/{organisation}/offerings > Create offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings"
		setBodyValue(&t.Request.Body, "title", dreddP2P)
		setBodyValue(&t.Request.Body, "organisation_id", orgUUID)
	})

	h.After("P2P/Offerings > p2p/api/organisations/{organisation}/offerings > Create offering", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		createdUUID = getBodyValue(&t.Real.Body, "id")
		if len(createdUUID) == 0 {
			t.Fail = "Unable to save the uuid of the created offering"
			return
		}
	})

	h.Before("P2P/Offerings > p2p/api/organisations/{organisation}/offerings > Retrieve all offerings", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings"
	})

	h.After("P2P/Offerings > p2p/api/organisations/{organisation}/offerings > Retrieve all offerings", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		// verify that created offerings are present in 'p2p/offerings' api call
		found := bytes.Contains([]byte(t.Real.Body), []byte(dreddP2P))
		if found {
			return
		}

		t.Fail = "Pre-created offering is missing"
	})

	h.Before("P2P/Offerings > p2p/api/organisations/{organisation}/offerings/{offering} > Retrieve offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
	})

	h.Before("P2P/Offerings > p2p/api/organisations/{organisation}/offerings/{offering} > Update offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID

		setBodyValue(&t.Request.Body, "organisation_id", orgUUID)
	})

	h.Before("P2P/Offerings > p2p/api/organisations/{organisation}/offerings/{offering} > Delete offering", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(createdUUID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + createdUUID
	})

	h.Before("P2P/Offering Media > p2p/api/organisations/{organisation}/offerings/{offering}/media/upload > Upload offering media", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(offeringID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/upload"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/upload"
	})

	h.After("P2P/Offering Media > p2p/api/organisations/{organisation}/offerings/{offering}/media/upload > Upload offering media", func(t *trans.Transaction) {
		// happens when api is down
		if t.Real == nil {
			return
		}

		mediaUUID = getBodyValue(&t.Real.Body, "id")
		if len(mediaUUID) == 0 {
			t.Fail = "Unable to save offering media ID"
			return
		}
	})

	h.Before("P2P/Offering Media > p2p/api/organisations/{organisation}/offerings/{offering}/media/ordering > Upload offering media order", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(offeringID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		type ordering struct {
			MediaID string `json:"media_id"`
			Index   int32  `json:"index"`
		}

		body := make([]ordering, 0)
		body = append(body, ordering{MediaID: mediaUUID, Index: 300})

		b, err := json.Marshal(body)
		if err != nil {
			return
		}

		t.Request.Body = string(b)

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/ordering"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/ordering"
	})

	h.Before("P2P/Offering Media > p2p/api/organisations/{organisation}/offerings/{offering}/media/{media} > Update offering media", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(offeringID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}
		if len(mediaUUID) == 0 {
			t.Fail = "Created offering media UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/" + mediaUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/" + mediaUUID
	})

	h.Before("P2P/Offering Media > p2p/api/organisations/{organisation}/offerings/{offering}/media/{media} > Delete offering media", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(offeringID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}
		if len(mediaUUID) == 0 {
			t.Fail = "Created offering media UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/" + mediaUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/offerings/" + offeringID + "/media/" + mediaUUID
	})

	h.Before("Trading/Offering Media > invest/api/offerings/{offering}/media > Get offering media", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(offeringID) == 0 {
			t.Fail = "Created offering UUID missing"
			return
		}

		t.Request.URI = "/invest/api/offerings/" + offeringID + "/media"
		t.FullPath = "/invest/api/offerings/" + offeringID + "/media"
	})

	h.Before("P2P/Invitations > p2p/api/organisations/{organisation}/invitations > Send invitation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}

		setBodyValue(&t.Request.Body, "name", dredd3)

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/invitations"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/invitations"
	})

	h.After("P2P/Invitations > p2p/api/organisations/{organisation}/invitations > Send invitation", func(t *trans.Transaction) {

		// happens when api is down
		if t.Real == nil {
			return
		}

		invitationUUID = getBodyValue(&t.Real.Body, "uuid")
		if len(invitationUUID) == 0 {
			t.Fail = "Unable to save invitation UUID"
			return
		}
		invitationCode = getBodyValue(&t.Real.Body, "code")
		if len(invitationCode) == 0 {
			t.Fail = "Unable to save invitation code"
			return
		}
	})

	h.Before("P2P/Invitations > p2p/api/organisations/{organisation}/invitations > Retrieve invitations", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/invitations"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/invitations"
	})

	h.Before("P2P/Invitations > p2p/api/organisations/{organisation}/invitations/{user} > Delete invitation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(invitationUUID) == 0 {
			t.Fail = "Invitation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/invitations/" + invitationUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/invitations/" + invitationUUID
	})

	h.Before("Trading/Invitations > invest/api/users/accept-invitation > Accept invitation", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}

		if len(invitationCode) == 0 {
			t.Fail = "Created invitation code is empty"
			return
		}
		if len(invitedUserUUID) == 0 {
			t.Fail = "Created invited user UUID is empty"
			return
		}

		// undelete the invitation
		orgUser := &models.OrganisationUser{
			UserID: invitedUserUUID,
			Status: models.OrganisationUserStatusInvited,
		}
		dbClient.Unscoped().Model(&models.OrganisationUser{}).Where(orgUser).Update("deleted_at", nil)

		setBodyValue(&t.Request.Body, "invitation_id", invitationCode)
	})

	h.Before("P2P/OrganisationUsers > p2p/api/organisations/{organisation}/users > Retrieve organisation users", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/users"
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/users"
	})

	h.Before("P2P/OrganisationUsers > p2p/api/organisations/{organisation}/users/{user} > Change organisation user role", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(dredd4.ID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
	})

	h.Before("P2P/OrganisationUsers > p2p/api/organisations/{organisation}/users/{user} > Delete organisation user", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(dredd4.ID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
	})

	h.Before("P2P/OrganisationUsers > p2p/api/organisations/{organisation}/users/{user} > Add organisation user", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(dredd4.ID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/users/" + dredd4.ID
	})

	server.Serve()
	defer server.Listener.Close()
}

func setBodyValue(body *string, key, value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	bodyMap[key] = value
	b, err := json.Marshal(bodyMap)
	if err != nil {
		return
	}

	*body = string(b)
}

func getBodyValue(body *string, key string) (value string) {

	if body == nil {
		return
	}

	bodyMap := map[string]interface{}{}

	err := json.Unmarshal([]byte(*body), &bodyMap)
	if err != nil {
		return
	}

	v, ok := bodyMap[key]
	if ok {
		// make sure it's a string
		vs, ok := v.(string)
		if ok {
			value = vs
		}
	}

	return
}
