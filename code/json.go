package btcvolatility

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"net/http"
)

func allHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	object, err := json.Marshal(d.Data[29:])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(object)
}

func latestHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	object, err := json.Marshal(d.Data[len(d.Data)-1])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(object)
}
