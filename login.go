package zteScanner

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

var jar, _ = cookiejar.New(nil)
var client = &http.Client{
	Jar: jar,
}

func getSessionToken(baseURL string) (string, error) {
	var responseData struct {
		SessionToken string `json:"sess_token"`
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?_type=loginData&_tag=login_entry", baseURL), nil)

	if err != nil {
		return "", err
	}

	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("error getting session token: %v", err)
	}
	u, _ := url.Parse(baseURL)

	// set cookies
	client.Jar.SetCookies(u, resp.Cookies())

	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&responseData)

	if err != nil {
		return "", fmt.Errorf("error reading session token: %v", err)
	}

	if responseData.SessionToken == "" {
		return "", fmt.Errorf("session token is empty")
	}
	return responseData.SessionToken, nil
}

func getLoginToken(baseURL string) (string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/?_type=loginData&_tag=login_token&_=%d", baseURL, time.Now().UnixMilli()), nil)

	if err != nil {
		return "", err

	}

	resp, err := client.Do(req)

	if err != nil {
		return "", fmt.Errorf("error getting login token: %v", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading login token: %v", err)

	}
	loginTokenXML := string(body)
	loginToken := strings.TrimLeft(strings.TrimRight(loginTokenXML, "</ajax_response_xml_root>"), "<ajax_response_xml_root>")

	if err != nil {
		return "", fmt.Errorf("error decoding login token xml: %v", err)
	}

	if loginToken == "" {
		return "", fmt.Errorf("error: empty login token")
	}

	return loginToken, nil
}

func encodePassword(password, loginToken string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password+loginToken)))
}

func forceLogin(baseURL, username, password string) error {

	sessionToken, err := getSessionToken(baseURL)
	if err != nil {
		return err
	}

	loginToken, err := getLoginToken(baseURL)
	if err != nil {
		return err
	}

	err = login(baseURL, username, encodePassword(password, loginToken), sessionToken)
	if err != nil {
		return err
	}

	payload := url.Values{}
	payload.Set("preempt_sessid", "random")
	payload.Set("action", "preempt")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/?_type=loginData&_tag=login_preempt", baseURL), strings.NewReader(payload.Encode()))

	if err != nil {
		return fmt.Errorf("error creating logout request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making logout request: %w", err)
	}

	defer resp.Body.Close()

	u, _ := url.Parse(baseURL)

	cookieSIDCount := 0
	for _, c := range resp.Cookies() {
		if c.Name == "SID" {
			cookieSIDCount += 1
		}

		if cookieSIDCount == 2 {
			client.Jar.SetCookies(u, []*http.Cookie{c})
			break
		}

	}

	return nil

}

// login and returns cookie
func login(baseURL, username, password, sessionToken string) error {

	payload := url.Values{}
	payload.Set("action", "login")
	payload.Set("Password", password)
	payload.Set("Username", username)
	payload.Set("_sessionTOKEN", sessionToken)

	var loginResponse struct {
		NeedRefresh bool   `json:"login_need_refresh"`
		ErrorMsg    string `json:"loginErrMsg"`
	}

	u, _ := url.Parse(baseURL)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/?_type=loginData&_tag=login_entry", baseURL), strings.NewReader(payload.Encode()))

	if err != nil {
		return fmt.Errorf("error creating login request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making login request: %w", err)
	}

	err = json.NewDecoder(resp.Body).Decode(&loginResponse)

	if err != nil {
		return fmt.Errorf("error decofding login response: %w", err)
	}

	// login unsuccessfull
	if !loginResponse.NeedRefresh {
		return fmt.Errorf("error: cannot login: %v", loginResponse.ErrorMsg)
	}

	// update new cookie
	cookieSIDCount := 0
	for _, c := range resp.Cookies() {
		if c.Name == "SID" {
			cookieSIDCount += 1
		}

		if cookieSIDCount == 2 {
			client.Jar.SetCookies(u, []*http.Cookie{c})
			break
		}

	}

	return nil

}
