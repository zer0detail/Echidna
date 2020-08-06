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
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        1024,
			MaxIdleConnsPerHost: 1024,
		},
	}
}

// ReqWorker is the worker responsible for sending requests to a remote site and retrieving the body
func ReqWorker(ctx context.Context, workerID int, reqCh chan string, resultsCh chan []byte, errCh chan error, client HTTPClient) {
	for req := range reqCh {
		body, err := SendRequest(ctx, client, req)
		if err != nil {
			errCh <- err
		}
		resultsCh <- body
	}
}

// PluginReq is a small struct to pass needed plugin info to the download worker
type PluginReq struct {
	URI      string
	FilePath string
	Index    int
}

// DownloadWorker is the worker responsible for retrieving requests and downloading zips
func DownloadWorker(ctx context.Context, workerID int, DownloadQueue chan PluginReq, ScanQueue chan int, errCh chan error, client HTTPClient) {
	refreshCounter := 0
	reqCounter := 0
	for req := range DownloadQueue {
		select {
		case <-ctx.Done():
			return
		default:
			reqCounter++
			err := download(ctx, client, req.FilePath, req.URI)
			if err != nil {
				//errCh <- err
				refreshCounter++
				//fmt.Printf("Worker %d: refreshing client. refreshCount: %d  reqCount: %d\n", workerID, refreshCounter, reqCounter)
				client = NewHTTPClient()
				continue
			}
			ScanQueue <- req.Index
		}

	}
}

// func measureRequest(start time.Time, uri string) {
// 	fmt.Printf("%.2f time taken for request to %s\n", time.Since(start).Seconds(), uri)
// }

// SendRequest sends a get request to an arbitrary site and returns the body
func SendRequest(ctx context.Context, client HTTPClient, uri string) ([]byte, error) {

	//defer measureRequest(time.Now(), uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("request.go:SendRequest() NewRequestWithContext() has failed with error: %s", err)
	}
	req.Header.Set("User-Agent", "Echidna V1.0")

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request.go:SendRequest() failed performing client.Do() with error: %s", err)
	}
	defer res.Body.Close()

	// Read out res.body into a var and create a new reader because res.body was hitting the client timeout
	// before we could read it. This uses response body faster to prevent hitting timeouts.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("request.go:SendRequest() failed to read response body with error: %s", err)
	}

	return body, nil

}

// download retrieves a remote file and stores it in the specified filepath
func download(ctx context.Context, client HTTPClient, filepath string, uri string) error {

	body, err := SendRequest(ctx, client, uri)
	if err != nil {
		return fmt.Errorf("request.go:Download() Failed to SendRequest for %s with error: %s", uri, err)
	}

	bodyReader := bytes.NewReader(body)

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("request.go:Download() Failed to create file with os.create() for %s with error: %s", filepath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, bodyReader)
	if err != nil {
		return fmt.Errorf("request.go:Download() Failed to write bytes to file for %s with error: %s", filepath, err)
	}

	return nil
}
