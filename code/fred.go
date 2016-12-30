package btcvolatility

import (
	"encoding/json"
	"github.com/GaryBoone/GoStats/stats"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"math"
	"sort"
	"strconv"
	"time"
)

const fredapikey string = "b19765ec1be9b52c32c2be76afc419ed"
const dateformat = "2006-01-02"

type FredResponse struct {
	Observations []FredObservation
}

type FredObservation struct {
	Date, Value string
}

func urlForSeries(z string) string {
	return "https://api.stlouisfed.org/fred/series/observations?series_id=" + z + "&api_key=" + fredapikey + "&file_type=json&observation_start=2010-07-18"
}

var updateSeries = delay.Func("Fred", func(c context.Context, z string) {

	u := urlForSeries(z)
	body, err := fetch(u, c)
	if err != nil {
		panic(err)
	}
	var response FredResponse
	if err = json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	now := time.Now()
	starttime, _ := time.Parse(dateformat, "2010-07-18")
	datalist := make(map[string]float64)

	for day := starttime; day.Before(now); day = day.AddDate(0, 0, 1) {
		datalist[day.Format(dateformat)] = math.NaN()
	}

	for _, obs := range response.Observations {
		if value, err := strconv.ParseFloat(obs.Value, 64); err != nil {
			// c.Infof("error: %v",err)
		} else {
			datalist[obs.Date] = value
		}
	}

	var data DataSet
	for k, v := range datalist {
		l := math.Log(v)
		d := DataPoint{
			Date:     k,
			Price:    v,
			Logprice: l,
		}
		data = append(data, d)
	}

	sort.Sort(DataSet(data))

	// ideally, we'd use multiple imputation to fill in weekends and holidays
	// must find Go package for multiple imputation

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
			if !math.IsNaN(point.Return) {
				returns = append(returns, point.Return)
			}
		}
		data[j].Volatility = stats.StatsSampleStandardDeviation(returns) * 100.0
		j += 1
	}
	k := 59
	for k < len(data) {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			if !math.IsNaN(point.Return) {
				returns = append(returns, point.Return)
			}
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
		k += 1
	}
	d := StoredDataSet{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", z, 0, nil), &d); err != nil {
		panic(err)
	}
})
