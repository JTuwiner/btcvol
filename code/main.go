package btcvolatility

import (
	"appengine"
	"appengine/datastore"
	"html/template"
	"net/http"
)

type RenderData struct {
	Latest30 float64
	Latest60 float64
	Data     DataSet
}

func init() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/all", allHandler)
	http.HandleFunc("/latest", latestHandler)
	http.HandleFunc("/csv", csvHandler)
}

var t = template.Must(template.New("content.html").ParseFiles(
	"templates/content.html",
	"templates/header.html",
	"templates/footer.html",
))

func mainHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "GET" || r.URL.Path != "/" {
		http.Redirect(w, r, "/", 303)
	}
	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var data RenderData
	data.Data = d.Data[29:]
	data.Latest30 = d.Data[len(d.Data)-1].Volatility
	data.Latest60 = d.Data[len(d.Data)-1].Volatility60
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
