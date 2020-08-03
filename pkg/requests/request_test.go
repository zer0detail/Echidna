package requests

//  go test -coverprofile="coverage.out"
//  go tool cover -html="coverage.out"

import (
	"context"
	"reflect"
	"strings"
	"testing"

	echidnatesting "github.com/Paraflare/Echidna/test"
)

// throwing away cancelfunc for now. add again when testing context cancelling
var ctx, _ = context.WithCancel(context.Background())

var (
	//dir, _ = os.Getwd()
	//sep    = string(os.PathSeparator)
	//outpath  = dir + sep + ".." + sep + ".." + sep + ".test" + sep + "test.zip"
	dummyURI = "http://127.0.0.1:8080"
	client   = &echidnatesting.MockClient{}
)

// func TestConsts(t *testing.T) {
// 	wantIdleConns := 20
// 	if maxIdleConnections != wantIdleConns {
// 		t.Errorf("maxIdleConnections constant is not expected value. Got: %d, want: %d\n", maxIdleConnections, wantIdleConns)
// 	}
// }

func TestCreateHTTPClient(t *testing.T) {

	testClient := NewHTTPClient()
	retType := reflect.TypeOf(testClient)
	if retType.String() != "*http.Client" {
		t.Errorf("CreateHTTPCLient() returned %s instead of *http.Client\n", retType)
	}
}

func TestSendRequestSuccess(t *testing.T) {

	client.SetReply(200, "Success", "")
	body, err := SendRequest(ctx, client, dummyURI)
	if err != nil {
		t.Errorf("Mocked SendRequest(%s) failed with error\n%s", dummyURI, err)
	}
	want := "Success"
	if !(strings.Contains(string(body), want)) {
		t.Errorf("Mocked SendRequest() did not return expected body.\nExpected: %s\nGot: %s", want, string(body))
	}
}

func TestSendRequestPageNotFound(t *testing.T) {

	client.SetReply(404, "This will 404", "")
	// Make sure SendRequest returns a protocol error when sending this malformed URI
	//expectedErr := fmt.Sprintf("Received non 200 StatusCode in SendRequest().\nStatusCode: %d", 404)
	_, err := SendRequest(ctx, client, dummyURI)
	if err != nil {
		if !(strings.Contains(err.Error(), "non 200 StatusCode")) {
			t.Errorf("Mocked SendRequest(%s) StatusCode check with error\n%s\nExpected: %s\n", dummyURI, err, "test") //expectedErr)
		}
	}

}

func TestSendRequestDoError(t *testing.T) {

	client.SetReply(404, "", "Page Not Found")

	_, err := SendRequest(ctx, client, dummyURI)
	if !(strings.Contains(err.Error(), "Page Not Found")) {
		t.Errorf("Mocked SendRequest(%s) failed\nReceived: %s\nExpected: Page Not Found\n", dummyURI, err)
	}
}

func TestSendRequestGoAway(t *testing.T) {

	client.SetReply(200, "", "GOAWAY")
	// Make sure SendRequest returns a protocol error when sending this malformed URI

	_, err := SendRequest(ctx, client, dummyURI)
	if !(strings.Contains(err.Error(), "GOAWAY")) {
		t.Errorf("Mocked SendRequest(%s) failed \nReceived: %s\nExpected: GOAWAY\n", dummyURI, err)
	}
}

// Commented out until I figure out mocking requirements for Download()

// func TestDownload(t *testing.T) {

// 	// Download askismet plugin since its pretty popular and likely to be downloadable
// 	// for atleast the near future.
// 	Download(outpath, "https://downloads.wordpress.org/plugin/akismet.4.1.6.zip")
// 	_, err := os.Stat(outpath)
// 	if os.IsNotExist(err) {
// 		t.Errorf("Failed to download plugin to directory %s in TestDownload()\n", outpath)
// 	}

// 	err = os.Remove(outpath)
// 	if err != nil {
// 		t.Errorf("TestDownload() test failed to remove test file:\n%s\n", outpath)
// 	}

// }
