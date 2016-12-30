package btcvolatility

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
)

func fetch(url string, c context.Context) ([]byte, error) {
	transport := urlfetch.Transport{
		Context: c,
		AllowInvalidServerCertificate: true,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// client := urlfetch.Client(c)
	// resp, err := client.Get(url)
	// if err != nil {
	// 	return nil, err
	// }
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return body, nil
}
