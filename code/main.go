package btcvolatility

import (
	"appengine"
	// "appengine/datastore"
	"html/template"
	"net/http"
)

type RenderData struct {
	Latest30, Latest60                                                         float64
	Data, Ether, GOLDAMGBD228NLBM, DEXUSEU, DEXBZUS, DEXCHUS, DEXTHUS, DEXJPUS BasicDataSet `datastore:",noindex"`
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

var renderdata RenderData
var bitcoindata []DataPoint
var latest DataPoint

func mainHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if r.Method != "GET" || r.URL.Path != "/" {
		http.Redirect(w, r, "/", 303)
	}
	if len(renderdata.Data) == 0 {
		preprocess(c)
	}
	if err := t.Execute(w, renderdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
