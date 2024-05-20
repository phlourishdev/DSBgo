package DSBgo

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const BaseURL = "https://mobileapi.dsbcontrol.de"

// Plan structure fo JSON respond from API
type Plan struct {
	Childs []Child // Not a typo
}

// Child structure of JSON respond from API
type Child struct {
	Id      string
	ConType int
	Date    string
	Title   string
	Detail  string
	Preview string
}

// ProcessedPlan as target structure for processed plans
type ProcessedPlan struct {
	ID           string
	IsHTML       bool
	UploadedDate string
	Title        string
	URL          string
	PreviewURL   string
}

// Authenticate retrieves the user token from DSB
func Authenticate(username, password string) (string, error) {
	// create map for GET request parameters
	paramsMap := map[string]string{
		"bundleid":   "de.heinekingmedia.dsbmobile",
		"appversion": "35",
		"osversion":  "22",
		"pushid":     "",
		"user":       username,
		"password":   password,
	}

	// convert map to url.Values
	params := url.Values{}
	for k, v := range paramsMap {
		params.Set(k, v)
	}

	fullURL := BaseURL + "/authid?" + params.Encode()

	// Perform the GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Convert the body to string
	token := string(body)

	// Check if token is empty and therefore if authentication has succeeded
	if token == "\"\"" {
		return "", errors.New("wrong user credentials")
	} else {
		token = strings.Trim(token, "\"")
	}

	return token, nil
}

// GetPlans fetches the information about all current substitute plans
func GetPlans(token string) ([]ProcessedPlan, error) {
	// Set parameter for GET request
	params := url.Values{}
	params.Set("authid", token)

	// Perform the GET request
	resp, err := http.Get(BaseURL + "/dsbtimetables?" + params.Encode())
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert JSON response to Plan[] struct type
	var rawPlans []Plan
	err = json.Unmarshal(body, &rawPlans)
	if err != nil {
		return nil, err
	}

	var plans []ProcessedPlan
	// Cycle through every entry in rawPlans
	for _, plan := range rawPlans {
		// Get data of every Child of rawPlans
		for _, child := range plan.Childs {
			// Create newChild var to store read data
			newChild := ProcessedPlan{
				ID:           child.Id,
				IsHTML:       child.ConType == 6,
				UploadedDate: child.Date,
				Title:        child.Title,
				URL:          child.Detail,
				PreviewURL:   "https://light.dsbcontrol.de/DSBlightWebsite/Data/" + child.Preview,
			}
			// Add to processed plans
			plans = append(plans, newChild)
		}
	}
	return plans, nil
}
