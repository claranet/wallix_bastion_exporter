package wallix

import (
	"context"
	"encoding/json"
	"fmt"
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

// Asbtracts any requests to Wallix bastion API.
func DoRequest(
	client *http.Client, method string, uri string, params map[string]string, basicAuth *BasicAuth,
) (results []map[string]interface{}, headers http.Header, err error) {
	req, err := http.NewRequestWithContext(context.Background(), method, uri, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create request to Wallix bastion %s: %w", uri, err)
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
		return nil, nil, fmt.Errorf("cannot do request to Wallix bastion %s: %w", uri, err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	// Authentication successful, stop here
	if res.StatusCode == http.StatusNoContent {
		return nil, res.Header, nil
	}

	decoder := json.NewDecoder(res.Body)

	if res.StatusCode != http.StatusOK {
		responseError := APIError{}
		if err := decoder.Decode(&responseError); err != nil {
			return nil, nil, fmt.Errorf("response http status not ok: %d, cannot decode error response: %w", res.StatusCode, err)
		}

		return nil, nil, fmt.Errorf("response http status not ok: %d, error response: %v", res.StatusCode, responseError)
	}

	if err := decoder.Decode(&results); err != nil {
		return nil, nil, fmt.Errorf("cannot decode results response %w", err)
	}

	return results, res.Header, nil
}

// Get users from /users API.
func GetUsers(client *http.Client, url string) (users []map[string]interface{}, err error) {
	users, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/users",
		map[string]string{
			"limit":  "-1",
			"fields": "user_name",
		},
		nil,
	)

	return users, err
}

// Get groups from /usergroups API.
func GetGroups(client *http.Client, url string) (groups []map[string]interface{}, err error) {
	groups, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/usergroups",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
		nil,
	)

	return groups, err
}

// Get devices from /devices API.
func GetDevices(client *http.Client, url string) (devices []map[string]interface{}, err error) {
	devices, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/devices",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
		nil,
	)

	return devices, err
}

// Get closed sesions for last sessionsClosedMinutes minutes.
func GetClosedSessions(
	client *http.Client, url string, sessionsClosedMinutes int,
) (sessionsClosed []map[string]interface{}, err error) {
	fromDate := time.Now().Add(
		-time.Minute * time.Duration(sessionsClosedMinutes),
	).Format(TimeFormat)

	sessionsClosed, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/sessions",
		map[string]string{
			"limit":      "-1",
			"fields":     "id",
			"date_field": "end",
			"status":     "closed",
			"from_date":  fromDate,
		},
		nil,
	)

	return sessionsClosed, err
}

// Get current active sesions from /sessions API.

func GetCurrentSessions(client *http.Client, url string) (sessionsCurrent []map[string]interface{}, err error) {
	sessionsCurrent, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/sessions",
		map[string]string{
			"limit":  "-1",
			"fields": "id",
			"status": "current",
		},
		nil,
	)

	return sessionsCurrent, err
}

// Get targets depdening on type from /targets API.
func GetTargets(client *http.Client, url string, targetType string) (targets []map[string]interface{}, err error) {
	targets, _, err = DoRequest(
		client,
		http.MethodGet,
		url+"/targets/"+targetType,
		map[string]string{
			"limit":  "-1",
			"fields": "id",
		},
		nil,
	)

	return targets, err
}
