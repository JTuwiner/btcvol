package btcvolatility

import (
	"encoding/json"
	"math"
	"net/http"

	"google.golang.org/appengine"
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

//add start
func allHandlerEther(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinEther) == 0 {
		preprocess(c)
	}

	i := 0
	for i < len(bitcoinEther) {
		if math.IsNaN(bitcoinEther[i].Volatility) {
			bitcoinEther[i].Volatility = 0
		}

		if math.IsNaN(bitcoinEther[i].Volatility60) {
			bitcoinEther[i].Volatility60 = 0
		}
		i += 1
	}

	object, err := json.Marshal(bitcoinEther)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerLTC(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinLTC) == 0 {
		preprocess(c)
	}

	i := 0
	for i < len(bitcoinLTC) {
		if math.IsNaN(bitcoinLTC[i].Volatility) {
			bitcoinLTC[i].Volatility = 0
		}

		if math.IsNaN(bitcoinLTC[i].Volatility60) {
			bitcoinLTC[i].Volatility60 = 0
		}
		i += 1
	}

	object, err := json.Marshal(bitcoinLTC)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerGOLDAMGBD228NLBM(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinGOLDAMGBD228NLBM) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinGOLDAMGBD228NLBM)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXUSEU(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXUSEU) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXUSEU)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXUSUK(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXUSUK) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXUSUK)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXBZUS(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXBZUS) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXBZUS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXCHUS(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXCHUS) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXCHUS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXTHUS(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXTHUS) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXTHUS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXJPUS(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXJPUS) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXJPUS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerDEXSFUS(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoinDEXSFUS) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoinDEXSFUS)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

//add end
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
