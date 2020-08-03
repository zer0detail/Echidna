// Package echidnatesting provides the exported MockClient for all tests that need to send web requests
package echidnatesting

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"net/http"
)

// MockClient lets us Mock the http.client's Do() func so we can do testing without making  requests
type MockClient struct {
	fakeStatus int
	fakeBody   string
	fakeError  string
}

// Do Satisfies the HTTPClient interface. Pseudo Do() function
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func (m *MockClient) doFunc(req *http.Request) (*http.Response, error) {

	body := ioutil.NopCloser(bytes.NewReader([]byte(m.fakeBody)))
	var err error
	if m.fakeError != "" {
		err = fmt.Errorf(m.fakeError)
	} else {
		err = nil
	}
	return &http.Response{
		StatusCode: m.fakeStatus,
		Body:       body,
	}, err
}

// SetReply lets us set the servers reply for the mocked request
func (m *MockClient) SetReply(resCode int, body, err string) {
	m.fakeStatus = resCode
	m.fakeBody = body
	m.fakeError = err
}

// DummyBody returns a byte array containing a fake wp plugin struct
// that can be unmarshaled
func (m *MockClient) DummyBody() string {
	basedir := build.Default.GOPATH
	fmt.Print(basedir)
	data, err := ioutil.ReadFile(basedir + "/src/github.com/zaptitude/echidna/test/dummydata.json")
	if err != nil {
		log.Fatal(err)
	}
	return string(data)

}
