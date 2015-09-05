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
	GOLDAMGBD228NLBM DataSet
	DEXUSEU  DataSet
	DEXBZUS  DataSet
	DEXCHUS  DataSet
	DEXTHUS  DataSet
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
	var gold StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "GOLDAMGBD228NLBM", 0, nil), &gold); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var euro StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSEU", 0, nil), &euro); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var brazil StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXBZUS", 0, nil), &brazil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var china StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXCHUS", 0, nil), &china); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var thailand StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXTHUS", 0, nil), &thailand); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var data RenderData
	data.Data = d.Data[29:]
	data.GOLDAMGBD228NLBM = gold.Data[29:]
	data.DEXUSEU = euro.Data[29:]
	data.DEXBZUS = brazil.Data[29:]
	data.DEXCHUS = china.Data[29:]
	data.DEXTHUS = thailand.Data[29:]
	data.Latest30 = d.Data[len(d.Data)-1].Volatility
	data.Latest60 = d.Data[len(d.Data)-1].Volatility60
	if err := t.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
