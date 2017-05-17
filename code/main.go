package btcvolatility

import (
	"html/template"
	"net/http"
	"time"

	"google.golang.org/appengine"
)

type RenderData struct {
	Latest30, Latest60                                                                           float64
	Data, Ether, GOLDAMGBD228NLBM, DEXUSEU, DEXUSUK, DEXBZUS, DEXCHUS, DEXTHUS, DEXJPUS, DEXSFUS BasicDataSet `datastore:",noindex"`
}

func init() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/debuge", debugeHandler)
	http.HandleFunc("/all", allHandler)
	//add start
	http.HandleFunc("/allLTC", allHandlerLTC)
	http.HandleFunc("/allEther", allHandlerEther)
	http.HandleFunc("/allGOLDAMGBD228NLBM", allHandlerGOLDAMGBD228NLBM)
	http.HandleFunc("/allDEXUSEU", allHandlerDEXUSEU)
	http.HandleFunc("/allDEXUSUK", allHandlerDEXUSUK)
	http.HandleFunc("/allDEXBZUS", allHandlerDEXBZUS)
	http.HandleFunc("/allDEXCHUS", allHandlerDEXCHUS)
	http.HandleFunc("/allDEXTHUS", allHandlerDEXTHUS)
	http.HandleFunc("/allDEXJPUS", allHandlerDEXJPUS)
	http.HandleFunc("/allDEXSFUS", allHandlerDEXSFUS)
	//add end
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
var bitcoinEther []DataPoint            //add
var bitcoinLTC []DataPoint              //add
var bitcoinGOLDAMGBD228NLBM []DataPoint //add
var bitcoinDEXUSEU []DataPoint          //add
var bitcoinDEXUSUK []DataPoint          //add
var bitcoinDEXBZUS []DataPoint          //add
var bitcoinDEXCHUS []DataPoint          //add
var bitcoinDEXTHUS []DataPoint          //add
var bitcoinDEXJPUS []DataPoint          //add
var bitcoinDEXSFUS []DataPoint          //add

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

func debugeHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	now := time.Now()
	u := "http://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
	body, err := fetch(u, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(body)
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := "https://etherchain.org/api/statistics/price"
	body, err := fetch(u, c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(body)
}
