package requests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTTPClient interface so we can mock http clients in testing
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTPClient for connection re-use
func NewHTTPClient() HTTPClient {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 60 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}
}

// SendRequest sends a get request to an arbitrary site and returns the body
func SendRequest(ctx context.Context, client HTTPClient, uri string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Echidna V1.0")
	res, err := client.Do(req)
	if err != nil {
		if strings.Contains(string(err.Error()), "GOAWAY") {
			return nil, fmt.Errorf("Error in sendrequest() performing client.Do() with error\n%s\nAttempting to refresh client", err)
		} else if strings.Contains(err.Error(), "context canceled") {
			// return nil nil for canceled requests
			return nil, nil
		} else {
			return nil, fmt.Errorf("Error in SendRequest performing client.Do() with error\n%s", err)
		}
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Received non 200 StatusCode in SendRequest().\nStatusCode: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

// Download retrieves a remote file and stores it in the specified filepath
func Download(ctx context.Context, client HTTPClient, filepath string, uri string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("NewRequest in download() has failed with error\n %s", err)
	}
	req.Header.Set("User-Agent", "Echidna V1.0")

	res, err := client.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "GOAWAY") {
			fmt.Println("Received GOAWAY, refreshing client")
			return nil
		} else if strings.Contains(err.Error(), "context canceled") {
			// return nil for canceled requests
			return nil
		} else {
			return fmt.Errorf("Error in SendRequest performing client.Do() with error\n%s", err)
		}
	}

	if res.Body != nil {
		defer res.Body.Close()
	} else {
		// log.WithFields(log.Fields{
		// 	"status": res.StatusCode,
		// 	"URI":    uri,
		// }).Warn("request.go:Download() response body is nil, Skipping download")
		return fmt.Errorf("request.go:Download() response body is nil")
	}
	// Read out res.body into a var and create a new reader because res.body was hitting the client timeout
	// before we could read it. This uses response body faster to prevent hitting timeouts.
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("request.go:Download() failed to read response body with error\t%s", err)
	}

	bodyReader := bytes.NewReader(bodyBytes)

	if res.StatusCode != 200 {
		// log.WithFields(log.Fields{
		// 	"status": res.StatusCode,
		// 	"URI":    uri,
		// }).Warn("request.go:Download() Did not receive 200 OK, Skipping download.")
		return fmt.Errorf("Received non 200 StatusCode in Download().\tStatusCode: %d", res.StatusCode)
	}
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Failed to create file with os.create() for %s with error\n%s", filepath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, bodyReader)
	if err != nil {
		return fmt.Errorf("Failed to write bytes to file for %s with error\n%s", filepath, err)
	}

	return nil
}
