package pages

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
)

// addExtraParams adds some more params to URL
func addExtraParams(baseURL string, extraParams url.Values) (string, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %v", err)
	}

	queryParams := parsedURL.Query()
	for key, values := range extraParams {
		queryParams[key] = append(queryParams[key], values...)
	}

	parsedURL.RawQuery = queryParams.Encode()
	return parsedURL.String(), nil
}

// postRequest handles our post requests
func postRequest(
	baseUrl string, requestData interface{},
	username string, password string) (*http.Response, error) {
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// Add basic authentication
	auth := username + ":" + password
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Set("Authorization", basicAuth)

	// Create custom transport with TLS settings
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			// Set any necessary TLS settings here
			// For example, to disable certificate validation:
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// postTEIsPayload ...
func postTEIsPayload(postURL string, payLoad []TrackedEntityInstance, username, password string) error {

	var teisPayload = TeisPayload{TrackedEntityInstances: payLoad}
	// Let's push the payload
	_, err := postRequest(postURL, teisPayload, username, password)
	if err != nil {
		log.Println("Error queuing chunk: ", err)
	}
	return err
}

// postTrackerPayload is a generic function to post any tracker payload
func postTrackerPayload(postURL string, payLoad interface{}, username, password string) error {
	switch t := payLoad.(type) {
	case []TrackedEntityInstance:
		fmt.Printf("Type of payload is %v \n", t)
		var teisPayload = TeisPayload{TrackedEntityInstances: payLoad.([]TrackedEntityInstance)}
		_, err := postRequest(postURL, teisPayload, username, password)
		if err != nil {
			log.Println("Error queuing Tracked Entities chunk: ", err)
		}
		return err
	case []Enrollment:
		fmt.Printf("Type of payload is %v \n", t)
		var enrollmentsPayload = EnrollmentsPayload{Enrollments: payLoad.([]Enrollment)}
		_, err := postRequest(postURL, enrollmentsPayload, username, password)
		if err != nil {
			log.Println("Error queuing enrollments chunk: ", err)
		}
		return err
	case []Event:
		fmt.Printf("Type of payload is %v \n", t)
		var eventsPayload = EventsPayload{Events: payLoad.([]Event)}
		_, err := postRequest(postURL, eventsPayload, username, password)
		if err != nil {
			log.Println("Error queuing events chunk: ", err)
		}
		return err
	default:
		log.Printf("Unsupported payload type: %v \n", t)
		return fmt.Errorf("Unsupported payload type: %v \n", t)
	}
	return nil
}
