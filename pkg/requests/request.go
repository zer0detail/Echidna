package requests

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}
}

// SendRequest sends a get request to an arbitrary site and returns the body
func SendRequest(ctx context.Context, client HTTPClient, uri string) ([]byte, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("\nNewRequestWithContext in SendRequest() has failed with error\n %s", err)
	}
	req.Header.Set("User-Agent", "Echidna V1.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("\nError in SendRequest() performing client.Do() with error\n%s", err)
	}
	defer res.Body.Close()

	// Read out res.body into a var and create a new reader because res.body was hitting the client timeout
	// before we could read it. This uses response body faster to prevent hitting timeouts.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("\nrequest.go:SendRequest() failed to read response body with error\t%s", err)
	}

	return body, nil

}

// Download retrieves a remote file and stores it in the specified filepath
func Download(ctx context.Context, client HTTPClient, filepath string, uri string) error {

	body, err := SendRequest(ctx, client, uri)
	if err != nil {
		return err
	}

	bodyReader := bytes.NewReader(body)

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
