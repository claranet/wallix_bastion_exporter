package wallix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	// Format expected by Wallix API on some resources like "sessions".
	TimeFormat = "2006-01-02 15:04:05"
)

// To pass credentials information to first request which login to API.
type BasicAuth struct {
	Username string
	Password string
}

type APIError struct {
	Error       string `json:"error"`
	Description string `json:"description"`
}

// Wraps any requests to Wallix bastion API.
func doRequest(
	client *http.Client, method string, url string, params map[string]string, basicAuth *BasicAuth,
) (body []byte, err error) {
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request to Wallix bastion %s: %w", url, err)
	}

	if params != nil {
		q := req.URL.Query()
		for k, v := range params {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if basicAuth != nil {
		req.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot do request to Wallix bastion %s: %w", url, err)
	}

	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}

	// Authentication successful, stop here
	if res.StatusCode == http.StatusNoContent {
		return body, nil
	}

	if res.StatusCode != http.StatusOK {
		statusError := fmt.Errorf("response http status not ok: %d", res.StatusCode)
		responseError := APIError{}
		if json.NewDecoder(res.Body).Decode(&responseError) == nil {
			return body, fmt.Errorf("%w, api error response: %v", statusError, responseError)
		}

		return body, fmt.Errorf("%w, plain text response: %s", statusError, string(body))
	}

	return body, nil
}

func QuerySchemes(
	client *http.Client, url string, params map[string]string,
) (results []map[string]interface{}, err error) {
	body, err := doRequest(
		client,
		http.MethodGet,
		url,
		params,
		nil,
	)
	if err != nil {
		return
	}
	reader := bytes.NewReader(body)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&results); err != nil {
		return nil, fmt.Errorf("cannot decode response as json list %w: %s", err, string(body))
	}

	return
}

func QueryScheme(
	client *http.Client, url string, params map[string]string,
) (result map[string]interface{}, err error) {
	body, err := doRequest(
		client,
		http.MethodGet,
		url,
		params,
		nil,
	)
	if err != nil {
		return
	}
	reader := bytes.NewReader(body)
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&result); err != nil {
		return nil, fmt.Errorf("cannot decode response as json object %w: %s", err, string(body))
	}

	return
}

// Authenticate on Wallix API to test or/and get cookie.
func Authenticate(
	client *http.Client, url string, user string, password string,
) (err error) {
	_, err = doRequest(
		client,
		http.MethodPost,
		url,
		nil,
		&BasicAuth{
			Username: user,
			Password: password,
		},
	)

	return
}

// Get users from /users API.
func GetUsers(client *http.Client, url string) (users []map[string]interface{}, err error) {
	users, err = QuerySchemes(
		client,
		url+"/users",
		map[string]string{
			"limit":  "-1",
			"fields": "user_name",
		},
	)

	return users, err
}

// Get groups from /usergroups API.
func GetGroups(client *http.Client, url string) (groups []map[string]interface{}, err error) {
	groups, err = QuerySchemes(
		client,
		url+"/usergroups",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
	)

	return groups, err
}

// Get devices from /devices API.
func GetDevices(client *http.Client, url string) (devices []map[string]interface{}, err error) {
	devices, err = QuerySchemes(
		client,
		url+"/devices",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
	)

	return devices, err
}

// Get closed sessions for last sessionsClosedMinutes minutes.
func GetClosedSessions(
	client *http.Client, url string, sessionsClosedMinutes int,
) (sessionsClosed []map[string]interface{}, err error) {
	fromDate := time.Now().Add(
		-time.Minute * time.Duration(sessionsClosedMinutes),
	).Format(TimeFormat)

	sessionsClosed, err = QuerySchemes(
		client,
		url+"/sessions",
		map[string]string{
			"limit":      "-1",
			"fields":     "id",
			"date_field": "end",
			"status":     "closed",
			"from_date":  fromDate,
		},
	)

	return sessionsClosed, err
}

// Get current active sessions from /sessions API.

func GetCurrentSessions(client *http.Client, url string) (sessionsCurrent []map[string]interface{}, err error) {
	sessionsCurrent, err = QuerySchemes(
		client,
		url+"/sessions",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
			"status": "current",
		},
	)

	return sessionsCurrent, err
}

// Get targets depdening on type from /targets API.
func GetTargets(client *http.Client, url string, targetType string) (targets []map[string]interface{}, err error) {
	targets, err = QuerySchemes(
		client,
		url+"/targets/"+targetType,
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
	)

	return targets, err
}

// Get encryption information from /encryption API.
func GetEncryption(client *http.Client, url string) (encryption map[string]interface{}, err error) {
	encryption, err = QueryScheme(
		client,
		url+"/encryption",
		nil,
	)

	return encryption, err
}

// Get license information from /licenseInfo API.
func GetLicense(client *http.Client, url string) (license map[string]interface{}, err error) {
	license, err = QueryScheme(
		client,
		url+"/licenseinfo",
		nil,
	)

	return license, err
}
