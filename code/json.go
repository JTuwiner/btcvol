package btcvolatility

import (
	"appengine"
	// "appengine/datastore"
	"encoding/json"
	"net/http"
)

func allHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoindata) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoindata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if latest.Volatility == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(latest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}
