package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func GetURL(address string, port int, endpoint string) string {
	return fmt.Sprintf("http://%s:%d%s", address, port, endpoint)
}

func ForwardRequest(originalRequest *http.Request, forwardURL string) error {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	forwardedRequest, err := http.NewRequest(originalRequest.Method, forwardURL, originalRequest.Body)
	if err != nil {
		return fmt.Errorf("Cannot create new Forward Request: %v", err.Error())
	}

	// Copy headers from the original request to the forwarded request
	for key, values := range originalRequest.Header {
		for _, value := range values {
			forwardedRequest.Header.Add(key, value)
		}
	}

	response, err := client.Do(forwardedRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	_, err = io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return nil
}

func SendPostRequest(url string, payload []byte) error {

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("Error creating request to %s: %v\n", url, err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Error sending POST request to %s: %v\n", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unsuccessfully request to %s: %s\n", url, resp.Status)
	}

	return nil
}

func SendGetRequest(ctx context.Context, baseURL string, queryParams map[string]string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Parse the base URL to append the query params
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("Error parsing URL %s: %v", baseURL, err)
	}

	// Add query parameters if provided
	if queryParams != nil {
		q := parsedURL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		parsedURL.RawQuery = q.Encode()
	}

	// Create the GET request
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating GET request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending GET request to %s: %v", parsedURL.String(), err)
	}

	return resp, nil
}
