package btcvolatility

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/GaryBoone/GoStats/stats"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	//"google.golang.org/appengine/log"
)

type CoinDeskResponse struct {
	Bpi map[string]float64
}

//add start
type LTCResponse struct {
	Price []LTCResponsePointArray
}

type LTCResponsePointArray struct {
	Kind []LTCResponsePoint
}

type LTCResponsePoint struct {
	Time time.Time
	Usd  float64
}

//end

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
	updateLTC.Call(c)
	updateSeries.Call(c, "GOLDAMGBD228NLBM")
	updateSeries.Call(c, "DEXUSEU")
	updateSeries.Call(c, "DEXUSUK")
	updateSeries.Call(c, "DEXBZUS")
	updateSeries.Call(c, "DEXCHUS")
	updateSeries.Call(c, "DEXTHUS")
	updateSeries.Call(c, "DEXJPUS")
	updateSeries.Call(c, "DEXSFUS")
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

var updateLTC = delay.Func("ltc", func(c context.Context) {
	u := "http://www.coincap.io/history/LTC"
	body, err := fetch(u, c)
	if err != nil {
		panic(err)
	}
	var f interface{}
	if err = json.Unmarshal(body, &f); err != nil {
		panic(err)
	}

	m := f.(map[string]interface{})
	var data DataSet
	datalist := make(map[string]float64)
	starttime, _ := time.Parse(dateformat, "2010-07-18")
	now := time.Now()

	for day := starttime; day.Before(now); day = day.AddDate(0, 0, 1) {
		datalist[day.Format(dateformat)] = math.NaN()
	}

	for k, v := range m {
		if k != "price" {
			continue
		}
		switch vv := v.(type) {
		case []interface{}:
			for _, u := range vv {
				switch vvv := u.(type) {
				case []interface{}:
					var date string
					for ii, uu := range vvv {
						if ii == 0 {
							datefloat := uu.(float64) / 1000
							dateint := int64(datefloat)
							date = time.Unix(dateint, 0).Format(dateformat)
							date = time.Unix(dateint, -5).Format(dateformat)
						} else {
							if math.IsNaN(datalist[date]) {
								datalist[date] = uu.(float64)
							}
						}
					}
				}
			}
		default:
			fmt.Println(k, "is of a type I don't know how to handle")
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

	fixnum := 0
	i := 1
	for i < len(data) {
		if math.IsNaN(data[i].Logprice) {
			data[i].Logprice = data[i-1].Logprice
			fixnum += 1
		}

		data[i].Return = data[i].Logprice - data[i-1].Logprice
		//log.Infof(c, "Date: %v USD: %v LogValue: %v ReturnValue: %v", data[i].Date, data[i].Price, data[i].Logprice, data[i].Return)
		i += 1
	}
	j := 29
	for j < (len(data) + fixnum) {
		subset := data[j-29 : j]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[j].Volatility = stats.StatsSampleStandardDeviation(returns) * 100.0
		//log.Infof(c, "Date: %v Value: %v Length: %v", data[j].Date, data[j].Volatility, returns)
		j += 1
	}
	k := 59
	for k < (len(data) + fixnum) {
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
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "ltc", 0, nil), &e); err != nil {
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
