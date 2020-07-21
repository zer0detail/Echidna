package requests

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

var (
	dir, _  = os.Getwd()
	sep     = string(os.PathSeparator)
	outpath = dir + sep + ".." + sep + ".." + sep + ".test" + sep + "test.zip"
)

// func TestConsts(t *testing.T) {
// 	wantIdleConns := 20
// 	if maxIdleConnections != wantIdleConns {
// 		t.Errorf("maxIdleConnections constant is not expected value. Got: %d, want: %d\n", maxIdleConnections, wantIdleConns)
// 	}
// }

func TestCreateHTTPClient(t *testing.T) {

	testClient := createHTTPClient()
	retType := reflect.TypeOf(testClient)
	if retType.String() != "*http.Client" {
		t.Errorf("CreateHTTPCLient() returned %s instead of *http.Client\n", retType)
	}
}

func TestSendRequest(t *testing.T) {
	exampleURI := "http://example.com"
	body, err := SendRequest(exampleURI)
	if err != nil {
		t.Errorf("SendRequest(%s) failed with error\n%s", exampleURI, err)
	}
	want := "<title>Example Domain</title>"
	if !(strings.Contains(string(body), want)) {
		t.Errorf("Body from example.com did not contain expected text.\nExpected: %s\nGot: %s", want, string(body))
	}
}

func TestBadURIs(t *testing.T) {
	// Make sure SendRequest returns a protocol error when sending this malformed URI
	exampleURI := "http/example.com"
	_, err := SendRequest(exampleURI)
	if !(strings.Contains(err.Error(), "unsupported protocol scheme")) {
		t.Errorf("SendRequest(%s) failed with error\n%s", exampleURI, err)
	}
	err = Download(outpath, exampleURI)
	if !(strings.Contains(err.Error(), "unsupported protocol scheme")) {
		t.Errorf("Download(%s) failed with error\n%s", exampleURI, err)
	}
}

func TestDownload(t *testing.T) {

	// Download askismet plugin since its pretty popular and likely to be downloadable
	// for atleast the near future.
	Download(outpath, "https://downloads.wordpress.org/plugin/akismet.4.1.6.zip")
	_, err := os.Stat(outpath)
	if os.IsNotExist(err) {
		t.Errorf("Failed to download plugin to directory %s in TestDownload()\n", outpath)
	}

	err = os.Remove(outpath)
	if err != nil {
		t.Errorf("TestDownload() test failed to remove test file:\n%s\n", outpath)
	}

}
