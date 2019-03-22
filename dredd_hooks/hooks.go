package main

import (
	"cig-exchange-libs/models"
	"encoding/json"
	"fmt"

	cigExchange "cig-exchange-libs"

	"github.com/lib/pq"
	"github.com/snikch/goodman/hooks"
	trans "github.com/snikch/goodman/transaction"
)

const dredd = "dredd"
const dredd2 = "dredd2"
const dredd3 = "dredd3"
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

	// save created record UUID here
	createdUUID := ""

	// prepare the database:
	// 1. delete 'dredd' users if it exists (first name  = 'dredd', 'dredd2', 'dredd3')
	// 2. delete 'dredd' organisations if it exists (reference key = 'dredd', 'dredd2', 'dredd3')
	// 3. create 'dredd' organisation (user will be registered with it)
	// 4. create some offerings belonging to 'dredd' organization
	// 5. verify that created offerings are present in 'invest/offerings' api call

	// delete 'dredd' users if it exists (first name  = 'dredd', 'dredd2', 'dredd3')
	dreddUsers := [3]string{dredd, dredd2, dredd3}
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

	// create 'dredd' organisation
	org := models.Organisation{
		Name:         dredd,
		ReferenceKey: dredd,
	}
	err := dbClient.Create(&org).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create organisation:")
		fmt.Println(err.Error())
	}
	orgUUID = org.ID

	// create some offerings belonging to 'dredd' organization
	offering := models.Offering{
		Title:          dredd,
		OrganisationID: orgUUID,
		Type:           make(pq.StringArray, 0),
		IsVisible:      true,
	}
	err = dbClient.Create(&offering).Error
	if err != nil {
		fmt.Println("ERROR: prepareDatabase: create offering:")
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
		offerings := make([]models.Offering, 0)
		err := json.Unmarshal([]byte(t.Real.Body), &offerings)
		if err != nil {
			t.Fail = fmt.Sprintf("Unable to parse response: %v", err.Error())
			return
		}

		for _, offering := range offerings {
			if offering.Title == dredd {
				// we found a match, api works fine
				return
			}
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

	h.Before("Trading/Users > invest/api/users/switch/{organisation} > Switch Organisation", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}

		t.Request.URI = "/invest/api/users/switch/" + orgUUID
		t.FullPath = "/invest/api/users/switch/" + orgUUID
	})

	h.Before("P2P/Organisation > p2p/api/organisations > Create organisation", func(t *trans.Transaction) {

		if t.Request == nil {
			return
		}

		setBodyValue(&t.Request.Body, "reference_key", dredd2)
		setBodyValue(&t.Request.Body, "name", dredd2)
	})

	h.Before("P2P/Organisation > p2p/api/organisations/{organisation} > Retrieve organisation", func(t *trans.Transaction) {
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

	h.Before("P2P/Organisation > p2p/api/organisations/{organisation} > Update organisation", func(t *trans.Transaction) {
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

	h.Before("P2P/Organisation > p2p/api/organisations/{organisation} > Delete organisation", func(t *trans.Transaction) {
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
		offerings := make([]models.Offering, 0)
		err := json.Unmarshal([]byte(t.Real.Body), &offerings)
		if err != nil {
			t.Fail = fmt.Sprintf("Unable to parse response: %v", err.Error())
			return
		}

		p2pFound := false
		homepageFound := false
		for _, offering := range offerings {
			if offering.Title == dredd {
				homepageFound = true
				continue
			}

			if offering.Title == dreddP2P {
				p2pFound = true
				continue
			}
		}

		if p2pFound && homepageFound {
			// we found a match, api works fine
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

	h.Before("P2P/OrganisationUsers > p2p/api/organisations/{organisation}/users/{user} > Delete organisation user", func(t *trans.Transaction) {
		if t.Request == nil {
			return
		}
		if len(orgUUID) == 0 {
			t.Fail = "Organisation UUID missing"
			return
		}
		if len(userUUID) == 0 {
			t.Fail = "User UUID missing"
			return
		}

		t.Request.URI = "/p2p/api/organisations/" + orgUUID + "/users/" + userUUID
		t.FullPath = "/p2p/api/organisations/" + orgUUID + "/users/" + userUUID
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
