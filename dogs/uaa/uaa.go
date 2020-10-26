package uaa

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
)

// Client is a client to the UAA api
// NOTE: client does not refresh the access token, for simplicicity timeout is 599s per default
type Client struct {
	baseURL  string
	username string
	password string
}

// NewClient returns a client
func NewClient(username, password, url string) *Client {
	return &Client{
		baseURL:  url,
		username: username,
		password: password,
	}
}

type createUserResponse struct {
	ID string `json:"id"`
}

// CreateUser creates a user on uaa and returns its Guid
func (c *Client) CreateUser(userName string) (string, error) {
	password := randomString(16)
	data := map[string]interface{}{
		"username": userName,
		"password": password,
		"name":     map[string]interface{}{},
		"emails": []map[string]interface{}{
			{
				"value":   userName,
				"primary": true,
			},
		},
	}

	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := c.doUaaRequest("POST", "/Users", bytes.NewBuffer(bodyBytes))

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	jsonResp := createUserResponse{}
	json.Unmarshal(body, &jsonResp)
	if err != nil {
		return "", err
	}

	return jsonResp.ID, nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(guid string) error {
	resp, err := c.doUaaRequest("DELETE", "/Users"+"/"+guid, nil)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New("Delete request returned unexpected status code '" + resp.Status + "'")
	}
	return nil
}

type user struct {
	ID       string `json:"id"`
	Username string `json:"userName"`
}

type userList struct {
	Resources []user `json:"resources"`
}

// GetUserGUID gets a guid for a username
func (c *Client) GetUserGUID(username string) (string, error) {
	resp, err := c.doUaaRequest("GET", "/Users?filter=username+eq+%22"+username+"%22", nil)
	if err != nil {
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	userListResponse := userList{}
	err = json.Unmarshal(body, &userListResponse)
	if err != nil {
		return "", err
	}
	return userListResponse.Resources[0].ID, nil
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (c *Client) getAccessToken() (string, error) {
	authURL := c.baseURL + "/oauth/token"
	data := url.Values{
		"username":     []string{c.username},
		"password":     []string{c.password},
		"grant_type":   []string{"password"},
		"scope":        []string{""},
		"token_format": []string{"opaque"},
	}
	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer([]byte(data.Encode())))
	req.Header.Add("Authorization", "Basic Y2Y6")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tokenResponse := accessTokenResponse{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func (c *Client) doUaaRequest(method, urlPath string, body io.Reader) (*http.Response, error) {
	requestURL := c.baseURL + urlPath

	token, err := c.getAccessToken()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, requestURL, body)
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, err
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789$!'|[](){}<>./?")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
