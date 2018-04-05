package DispatchAuth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

//var statusCode int

//var tokenErr error
//var auth0Token string
var auth0GrantType = "client_credentials"
var _authresp Auth0Response
var _authReq Auth0Request
var _ingroovesAuth0Endpoint string

//var reqObject string

//Auth0Request represent as Request structure
type Auth0Request struct {
	Auth0GrantType    string `json:"grant_type"`
	Auth0Audience     string `json:"audience"`
	Auth0ClientID     string `json:"client_id"`
	Auth0ClientSecret string `json:"client_secret"`
}

//Auth0Response represent as Response structure
type Auth0Response struct {
	Auth0ACCESSTOKEN string    `json:"access_token"`
	Auth0SCOPE       string    `json:"scope"`
	Auth0EXPIREIN    int       `json:"expires_in"`
	Auth0TOKENTYPE   string    `json:"token_type"`
	Auth0ValidiTill  time.Time `json:"validTill"`
}

//InitialiseRequest will initialise Auth0Request request
func InitialiseRequest(Auth0Audience string, Auth0ClientID string, Auth0ClientSecret string, IngroovesAuth0Endpoint string) {

	_authReq = Auth0Request{
		Auth0GrantType:    auth0GrantType,
		Auth0Audience:     Auth0Audience,
		Auth0ClientID:     Auth0ClientID,
		Auth0ClientSecret: Auth0ClientSecret,
	}
	_ingroovesAuth0Endpoint = IngroovesAuth0Endpoint

}

//AddToken will create Auth0 Token & add to Request object
func AddToken(reqObj *http.Request) (*http.Request, error) {
	var err error

	if _authReq.Auth0Audience == "" || _authReq.Auth0ClientID == "" || _authReq.Auth0ClientSecret == "" || _ingroovesAuth0Endpoint == "" {
		err = errors.New("Initialize request objects first")
		return nil, err
	}

	if _authresp.Auth0ACCESSTOKEN == "" || IsExpired(_authresp.Auth0ValidiTill) {
		_authresp, err = GenerateAuth0Token(_authReq, _ingroovesAuth0Endpoint)

		if err != nil {
			return reqObj, err
		} else {
			reqObj.Header.Set("Bearer", _authresp.Auth0ACCESSTOKEN)
			reqObj.Header.Set("Content-Type", "application/json")
		}

	} else if _authresp.Auth0ACCESSTOKEN != "" {
		reqObj.Header.Set("Bearer", _authresp.Auth0ACCESSTOKEN)
		reqObj.Header.Set("Content-Type", "application/json")
	}
	return reqObj, err
}

//GenerateAuth0Token method will create Token & return response
func GenerateAuth0Token(auth Auth0Request, IngroovesAuth0Endpoint string) (Auth0Response, error) {

	var response Auth0Response
	var err error

	url := IngroovesAuth0Endpoint
	authRequest := &Auth0Request{Auth0GrantType: auth.Auth0GrantType, Auth0Audience: auth.Auth0Audience, Auth0ClientID: auth.Auth0ClientID, Auth0ClientSecret: auth.Auth0ClientSecret}
	jsonString, err := json.Marshal(authRequest)
	if err != nil {
		return response, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonString))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		return response, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return response, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if IsSuccess(resp.StatusCode) {
		err = json.Unmarshal([]byte(string(body)), &response)
		if err != nil {
			return response, err
		}
		currentDate := time.Now()
		response.Auth0ValidiTill = currentDate.Add(time.Minute * time.Duration((response.Auth0EXPIREIN/60)-5))
	}

	defer resp.Body.Close()

	return response, err
}

//IsSuccess gives true if the status code is under 2xx Success code...
func IsSuccess(statusCode int) bool {
	success := false
	switch {
	case statusCode == http.StatusOK:
		fallthrough
	case statusCode == http.StatusCreated:
		fallthrough
	case statusCode == http.StatusAccepted:
		fallthrough
	case statusCode == http.StatusNonAuthoritativeInfo:
		fallthrough
	case statusCode == http.StatusNoContent:
		fallthrough
	case statusCode == http.StatusResetContent:
		fallthrough
	case statusCode == http.StatusPartialContent:
		fallthrough
	case statusCode == http.StatusMultiStatus:
		fallthrough
	case statusCode == http.StatusAlreadyReported:
		fallthrough
	case statusCode == http.StatusIMUsed:
		success = true
	}
	return success
}

//IsExpired will return true if Token got expired
func IsExpired(validity time.Time) bool {
	var timeDiff time.Duration
	currentDate := time.Now()
	timeDiff = validity.Sub(currentDate)
	minsDiff := int(timeDiff.Minutes())
	if minsDiff > 1 {
		return false
	}
	return true
}
