package btcvolatility

import (
	"appengine"
	"appengine/datastore"
	"fmt"
	"net/http"
)

func csvHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=btcvol.csv")
	fmt.Fprintln(w, "date,price,logprice,return,volatility,volatility60")
	for _, v := range d.Data[29:] {
		fmt.Fprintf(w, "%v,%v,%v,%v,%v\n", v.Date, v.Price, v.Logprice, v.Return, v.Volatility/100.0, v.Volatility60/100.0)
	}
}
