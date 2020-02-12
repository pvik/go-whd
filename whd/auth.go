package whd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const urn string = "/helpdesk/WebObjects/Helpdesk.woa/ra/"

type authType int

const (
	ApiKeyAuth     authType = 0
	SessionKeyAuth authType = 1
	PasswordAuth   authType = 2
)

type User struct {
	Name string
	Pass string
	Type authType
}

func WrapAuth(req *http.Request, user User) {

	q := req.URL.Query()

	switch user.Type {
	case PasswordAuth:
		q.Add("username", user.Name)
		q.Add("password", user.Pass)
	case SessionKeyAuth:
		q.Add("username", user.Name)
		q.Add("sessionKey", user.Pass)
	case ApiKeyAuth:
		q.Add("apiKey", user.Pass)
	}

	req.URL.RawQuery = q.Encode()
}

func GetSessionKey(uri string, user User) (string, error) {
	req, err := http.NewRequest("GET", uri+urn+"Session", nil)
	if err != nil {
		return "", err
	}

	WrapAuth(req, user)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return "", err
	}

	data, _ := ioutil.ReadAll(resp.Body)

	var dataMap map[string]interface{}
	if err := json.Unmarshal(data, &dataMap); err != nil {
		log.Println("error unmarshalling: ", err)
		return "", err
	}

	sessionKey, ok := dataMap["sessionKey"].(string)
	if !ok {
		log.Println("invalid sessionKey in map")
		return "", err
	}

	return sessionKey, nil
}

func TerminateSession(uri string, sessionKey string) error {
	req, err := http.NewRequest("DELETE", uri+urn+"Session", nil)
	if err != nil {
		return err
	}

	q := req.URL.Query()
	q.Add("sessionKey", sessionKey)
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("The HTTP request failed with error %s\n", err)
		return err
	}

	data, _ := ioutil.ReadAll(resp.Body)

	if string(data) == "OK" {
		return nil
	}

	return fmt.Errorf("Invalid response: %s", data)
}
