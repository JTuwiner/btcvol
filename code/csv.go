package btcvolatility

import (
	"fmt"
	"google.golang.org/appengine"
	"net/http"
)

func csvHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	if len(bitcoindata) == 0 {
		preprocess(c)
	}
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=btcvol.csv")
	fmt.Fprintln(w, "date,price,logprice,return,volatility,volatility60")
	for _, v := range bitcoindata {
		fmt.Fprintf(w, "%v,%v,%v,%v,%v,%v\n", v.Date, v.Price, v.Logprice, v.Return, v.Volatility/100.0, zerotonull(v.Volatility60/100.0))
	}
}

func zerotonull(n float64) string {
	if n == 0 {
		return ""
	} else {
		return fmt.Sprintf("%v", n)
	}
}
