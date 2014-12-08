package btcvolatility

import (
	"appengine"
	"appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"time"
)

func fetch(url string, c appengine.Context) ([]byte, error) {
	transport := urlfetch.Transport{
		Context:                       c,
		Deadline:                      time.Duration(20) * time.Second,
		AllowInvalidServerCertificate: false,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return body, nil
}