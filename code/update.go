package btcvolatility

import (
	"appengine"
	"appengine/datastore"
	"encoding/json"
	"fmt"
	"github.com/GaryBoone/GoStats/stats"
	"math"
	"net/http"
	"sort"
	"time"
)

type CoinDeskResponse struct {
	Bpi map[string]float64
}

type DataPoint struct {
	Date                    string
	Price, Logprice, Return float64 `json:"-" datastore:",noindex"`
	Volatility              float64 `datastore:",noindex"`
}

type DataSet []DataPoint

type StoredDataSet struct {
	Data DataSet
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	now := time.Now()
	u := "https://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
	body, err := fetch(u, c)
	if err != nil {
		panic(err)
	}
	var response CoinDeskResponse
	if err = json.Unmarshal(body, &response); err != nil {
		panic(err)
	}
	var data DataSet
	for k, v := range response.Bpi {
		l := math.Log(v)
		d := DataPoint{
			Date:     k,
			Price:    v,
			Logprice: l,
		}
		data = append(data, d)
	}
	sort.Sort(DataSet(data))

	i := 1
	for i < len(data) {
		data[i].Return = data[i].Logprice - data[i-1].Logprice
		i += 1
	}
	j := 29
	for j < len(data) {
		subset := data[j-29 : j]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[j].Volatility = stats.StatsSampleStandardDeviation(returns) * 100.0
		j += 1
	}

	d := StoredDataSet{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "OK")
}

// Sorting

func (s DataSet) Len() int {
	return len(s)
}
func (s DataSet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s DataSet) Less(i, j int) bool {
	return s[i].Date < s[j].Date
}