package btcvolatility

import (
	"encoding/json"
	"fmt"
	"github.com/GaryBoone/GoStats/stats"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"math"
	"net/http"
	"sort"
	"time"
)

type CoinDeskResponse struct {
	Bpi map[string]float64
}

type EtherChainResponse struct {
	Data []EtherChainPoint
}

type EtherChainPoint struct {
	Time time.Time
	Usd  float64
}

type DataPoint struct {
	Date                    string
	Price, Logprice, Return float64 `json:"-" datastore:",noindex"`
	Volatility              float64 `datastore:",noindex"`
	Volatility60            float64 `datastore:",noindex"`
}

type DataSet []DataPoint

type StoredDataSet struct {
	Data DataSet
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	updateBitcoin.Call(c)
	updateEther.Call(c)
	updateSeries.Call(c, "GOLDAMGBD228NLBM")
	updateSeries.Call(c, "DEXUSEU")
	updateSeries.Call(c, "DEXUSUK")
	updateSeries.Call(c, "DEXBZUS")
	updateSeries.Call(c, "DEXCHUS")
	updateSeries.Call(c, "DEXTHUS")
	updateSeries.Call(c, "DEXJPUS")
	delayedpreprocess.Call(c)
	fmt.Fprint(w, "OK")
}

var updateBitcoin = delay.Func("Bitcoin", func(c context.Context) {
	now := time.Now()
	u := "http://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
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
	k := 59
	for k < len(data) {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
		k += 1
	}

	d := StoredDataSet{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		panic(err)
	}
})

var updateEther = delay.Func("Ether", func(c context.Context) {
	u := "https://etherchain.org/api/statistics/price"
	body, err := fetch(u, c)
	if err != nil {
		panic(err)
	}
	var response EtherChainResponse
	if err = json.Unmarshal(body, &response); err != nil {
		panic(err)
	}
	var data DataSet
	datalist := make(map[string]float64)
	starttime, _ := time.Parse(dateformat, "2010-07-18")
	now := time.Now()

	for day := starttime; day.Before(now); day = day.AddDate(0, 0, 1) {
		datalist[day.Format(dateformat)] = math.NaN()
	}
	for _, obs := range response.Data {
		date := obs.Time.Format(dateformat)
		if math.IsNaN(datalist[date]) {
			datalist[date] = obs.Usd
		}
	}

	for date, usd := range datalist {
		l := math.Log(usd)
		d := DataPoint{
			Date:     date,
			Price:    usd,
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
	k := 59
	for k < len(data) {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
		k += 1
	}

	e := StoredDataSet{
		Data: data,
	}
	// log.Infof(c, "ether dataset: %v", e)
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "ether", 0, nil), &e); err != nil {
		panic(err)
	}
})

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
