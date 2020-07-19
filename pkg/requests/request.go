package requests

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const (
	maxIdleConnections int = 20
	//requestTimeout     int = -1
)

var client *http.Client

func init() {
	client = createHTTPClient()
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: maxIdleConnections,
		},
		Timeout: -1, //time.Duration(requestTimeout) * time.Second,
	}

	return client
}

// SendRequest sends a get request to an arbitrary site and returns the body
func SendRequest(uri string) ([]byte, error) {

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Echidna V1.0")

	res, err := client.Do(req)
	if err != nil {
		if strings.Contains(string(err.Error()), "GOAWAY") {
			fmt.Printf("Error in sendrequest() performing client.Do() with error\n%s\nAttempting to refresh client..\n", err)
			client = createHTTPClient()
			SendRequest(uri)
		} else {
			return nil, fmt.Errorf("Error in SendRequest performing client.Do() with error\n%s", err)
		}

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil

}

// Download retrieves a remote file and stores it in the specified filepath
func Download(filepath string, uri string) error {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return fmt.Errorf("NewRequest in download() has failed with error\n %s", err)
	}
	req.Header.Set("User-Agent", "Echidna V1.0")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do() in download() has failed with error\n %s", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("Failed to create file with os.create() for %s with error\n%s", filepath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		if strings.Contains(string(err.Error()), "GOAWAY") {
			fmt.Printf("Error in download() with error\n%s\nAttempting to refresh client..\n", err)
			client = createHTTPClient()
			Download(filepath, uri)
		}
		return fmt.Errorf("Failed to write bytes to file for %s with error\n%s", filepath, err)
	}

	return nil

}
