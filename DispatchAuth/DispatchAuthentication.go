package DispatchAuth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

var auth0GrantType = "client_credentials"
var _authReq auth0Request
var _authResp auth0Response
var _ingroovesAuth0Endpoint string

//auth0Request represent as Request structure.
type auth0Request struct {
	Auth0GrantType    string `json:"grant_type"`
	Auth0Audience     string `json:"audience"`
	Auth0ClientID     string `json:"client_id"`
	Auth0ClientSecret string `json:"client_secret"`
}

//auth0Response represent as Response structure.
type auth0Response struct {
	Auth0AccessToken string    `json:"access_token"`
	Auth0Scope       string    `json:"scope"`
	Auth0ExpireIn    int       `json:"expires_in"`
	Auth0TokenType   string    `json:"token_type"`
	Auth0ValidiTill  time.Time `json:"validTill"`
}

//InitialiseRequest will initialise auth0Request request.
func InitialiseRequest(Auth0Audience string, Auth0ClientID string, Auth0ClientSecret string, IngroovesAuth0Endpoint string) {

	_authReq = auth0Request{
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

	if _authResp.Auth0AccessToken == "" || isExpired(_authResp.Auth0ValidiTill) {
		_authResp, err = generateAuth0Token(_authReq, _ingroovesAuth0Endpoint)

		if err != nil {
			return reqObj, err
		} else {
			reqObj.Header.Set("Bearer", _authResp.Auth0AccessToken)			
		}

	} else if _authResp.Auth0AccessToken != "" {
		reqObj.Header.Set("Bearer", _authResp.Auth0AccessToken)		
	}
	return reqObj, err
}

//GenerateAuth0Token method will create Token & return response.
func generateAuth0Token(auth auth0Request, IngroovesAuth0Endpoint string) (auth0Response, error) {

	var response auth0Response
	var err error

	url := IngroovesAuth0Endpoint
	authRequest := &auth0Request{Auth0GrantType: auth.Auth0GrantType, Auth0Audience: auth.Auth0Audience, Auth0ClientID: auth.Auth0ClientID, Auth0ClientSecret: auth.Auth0ClientSecret}
	requestString, err := json.Marshal(authRequest)
	if err != nil {
		return response, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestString))
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
	if isSuccess(resp.StatusCode) {
		err = json.Unmarshal([]byte(string(body)), &response)
		if err != nil {
			return response, err
		}
		currentDate := time.Now()
		response.Auth0ValidiTill = currentDate.Add(time.Minute * time.Duration((response.Auth0ExpireIn/60)-5))
	} else {
		err = errors.New("Auth0 token not generated successfully")
	}

	defer resp.Body.Close()

	return response, err
}

//isSuccess gives true if the status code is under 2xx Success code.
func isSuccess(statusCode int) bool {
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

//isExpired will return true if Token got expired.
func isExpired(validity time.Time) bool {
	var timeDiff time.Duration
	currentDate := time.Now()
	timeDiff = validity.Sub(currentDate)
	minsDiff := int(timeDiff.Minutes())
	if minsDiff > 1 {
		return false
	}
	return true
}
