package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/GaryBoone/GoStats/stats"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
)

const fredapikey string = "b19765ec1be9b52c32c2be76afc419ed"
const dateformat = "2006-01-02"

type DataPoint struct {
	Date                    string
	Price, Logprice, Return float64 `json:"-" datastore:",noindex"`
	Volatility              float64 `datastore:",noindex"`
	Volatility60            float64 `datastore:",noindex"`
}
type FredObservation struct {
	Date, Value string
}
type BasicDataSet []DataPoint
type DataSet []DataPoint
type FredResponse struct {
	Observations []FredObservation
}
type RenderData struct {
	Latest30, Latest60                                                                           float64
	Data, Ether, GOLDAMGBD228NLBM, DEXUSEU, DEXUSUK, DEXBZUS, DEXCHUS, DEXTHUS, DEXJPUS, DEXSFUS BasicDataSet `datastore:",noindex"`
}
type StoredDataSet struct {
	Data DataSet
}
type CoinDeskResponse struct {
	Bpi map[string]float64
}
type EtherChainPoint struct {
	Time time.Time
	Usd  float64
}
type EtherChainResponse struct {
	Data []EtherChainPoint
}
type LatestStruct struct {
	Date         string
	Volatility   float64
	Volatility60 float64
	Price        float64
}
type DataPointPrice struct {
	Date  string
	Price float64 `datastore:",noindex"`
}
type DataSetPrice []DataPointPrice
type StoredDataSetPrice struct {
	Data DataSetPrice
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/debug", debugHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/all", allHandler)
	http.HandleFunc("/allBTCPrice", allHandlerPrice)
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
	http.HandleFunc("/latest", latestHandler)
	http.HandleFunc("/latest-block", latestBlockHandler)
	http.HandleFunc("/csv", csvHandler)

	// port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080"
	// 	log.Printf("Defaulting to port %s", port)
	// }

	// log.Printf("Listening on port %s", port)
	// if err := http.ListenAndServe(":"+port, nil); err != nil {
	// 	log.Fatal(err)
	// }
	appengine.Main()
}

func urlForSeries(z string) string {
	return "https://api.stlouisfed.org/fred/series/observations?series_id=" + z + "&api_key=" + fredapikey + "&file_type=json&observation_start=2010-07-18"
}

var delayedpreprocess = delay.Func("preprocess", preprocess)

var updateBitcoin = delay.Func("Bitcoin", func(c context.Context) {
	now := time.Now()
	u := "http://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
	body, err := fetch(u)
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

	starttime, _ := time.Parse(dateformat, "2010-07-18")
	for day := starttime; day.Before(now); day = day.AddDate(0, 0, 1) {
		isExist := false
		prevDay := day.AddDate(0, 0, -1)
		var prevDayData DataPoint
		for i := 0; i < len(data); i++ {
			if data[i].Date == prevDay.Format(dateformat) {
				prevDayData = data[i]
			}

			if data[i].Date == day.Format(dateformat) {
				isExist = true
				break
			}
		}

		if !isExist {
			d := DataPoint{
				Date:     day.Format(dateformat),
				Price:    prevDayData.Price,
				Logprice: prevDayData.Logprice,
			}
			data = append(data, d)
		}
	}

	sort.Sort(DataSet(data))

	for i := 1; i < len(data); i++ {
		data[i].Return = data[i].Logprice - data[i-1].Logprice
	}
	for j := 29; j < len(data); j++ {
		subset := data[j-29 : j]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[j].Volatility = stats.StatsSampleStandardDeviation(returns) * 100.0
	}
	for k := 59; k < len(data); k++ {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
	}

	d := StoredDataSet{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		panic(err)
	}
})

var updateBitcoinPrice = delay.Func("BitcoinPrice", func(c context.Context) {
	now := time.Now()
	u := "http://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
	body, err := fetch(u)
	if err != nil {
		panic(err)
	}
	var response CoinDeskResponse
	if err = json.Unmarshal(body, &response); err != nil {
		panic(err)
	}
	var data DataSetPrice
	for k, v := range response.Bpi {
		d := DataPointPrice{
			Date:  k,
			Price: v,
		}
		data = append(data, d)
	}
	sort.Sort(DataSetPrice(data))

	u = "http://api.coindesk.com/v1/bpi/currentprice.json"
	body, err = fetch(u)
	if err != nil {
		panic(err)
	}

	var dat map[string]interface{}
	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	//fmt.Println(dat)

	strs := dat["bpi"].(interface{})
	m := strs.(map[string]interface{})
	for k, v := range m {
		if k != "USD" {
			continue
		}
		switch vv := v.(type) {
		case interface{}:
			//fmt.Println(k, "is an interface:", vv)
			n := vv.(map[string]interface{})
			for kk, vvv := range n {
				if kk == "rate_float" {
					fmt.Println(kk, vvv.(float64))
					d := DataPointPrice{
						Date:  now.Format("2006-01-02"),
						Price: vvv.(float64),
					}
					data = append(data, d)
				}
			}
		}
	}

	d := StoredDataSetPrice{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSetPrice", "data", 0, nil), &d); err != nil {
		panic(err)
	}
})

var updateEther = delay.Func("Ether", func(c context.Context) {
	u := "https://etherchain.org/api/statistics/price"
	body, err := fetch(u)
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
		i++
	}
	j := 29
	for j < len(data) {
		subset := data[j-29 : j]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[j].Volatility = stats.StatsSampleStandardDeviation(returns) * 100.0
		j++
	}
	k := 59
	for k < len(data) {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
		k++
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
	body, err := fetch(u)
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
			fixnum++
		}

		data[i].Return = data[i].Logprice - data[i-1].Logprice
		//log.Infof(c, "Date: %v USD: %v LogValue: %v ReturnValue: %v", data[i].Date, data[i].Price, data[i].Logprice, data[i].Return)
		i++
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
		j++
	}
	k := 59
	for k < (len(data) + fixnum) {
		subset := data[k-59 : k]
		var returns []float64
		for _, point := range subset {
			returns = append(returns, point.Return)
		}
		data[k].Volatility60 = stats.StatsSampleStandardDeviation(returns) * 100.0
		k++
	}

	e := StoredDataSet{
		Data: data,
	}
	// log.Infof(c, "ether dataset: %v", e)
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", "ltc", 0, nil), &e); err != nil {
		panic(err)
	}
})

var updateSeries = delay.Func("Fred", func(c context.Context, z string) {
	u := urlForSeries(z)
	body, err := fetch(u)
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

	i := 1
	for i < len(data) {
		data[i].Return = data[i].Logprice - data[i-1].Logprice
		i++
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
		j++
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
		k++
	}
	d := StoredDataSet{
		Data: data,
	}
	if _, err := datastore.Put(c, datastore.NewKey(c, "StoredDataSet", z, 0, nil), &d); err != nil {
		panic(err)
	}
})

var t = template.Must(template.New("content.html").ParseFiles(
	"templates/content.html",
	"templates/header.html",
	"templates/footer.html",
))

var renderdata RenderData
var latest DataPoint
var latestPrice float64
var bitcoindata []DataPoint
var bitcoindataprice []DataPointPrice
var bitcoinEther []DataPoint
var bitcoinLTC []DataPoint
var bitcoinGOLDAMGBD228NLBM []DataPoint
var bitcoinDEXUSEU []DataPoint
var bitcoinDEXUSUK []DataPoint
var bitcoinDEXBZUS []DataPoint
var bitcoinDEXCHUS []DataPoint
var bitcoinDEXTHUS []DataPoint
var bitcoinDEXJPUS []DataPoint
var bitcoinDEXSFUS []DataPoint

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	c := r.Context()
	if len(renderdata.Data) == 0 {
		preprocess(c)
	}

	if err := t.Execute(w, renderdata); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func debugHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	u := "http://api.coindesk.com/v1/bpi/historical/close.json?start=2010-07-18&end=" + now.Format("2006-01-02")
	g, err := fetch(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write(g)
}

func csvHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
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

func allHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
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

func allHandlerPrice(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	if len(bitcoindataprice) == 0 {
		preprocess(c)
	}
	object, err := json.Marshal(bitcoindataprice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func allHandlerEther(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
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
		i++
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
	c := r.Context()
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
		i++
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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
	c := r.Context()
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

func latestHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	updateBitcoinPrice.Call(c)
	preprocess(c)
	lv := LatestStruct{latest.Date, latest.Volatility, latest.Volatility60, latestPrice}
	object, err := json.Marshal(lv)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func latestBlockHandler(w http.ResponseWriter, r *http.Request) {
	u := "https://blockchain.info/latestblock"
	body, err := fetch(u)
	if err != nil {
		panic(err)
	}

	var dat map[string]interface{}

	if err := json.Unmarshal(body, &dat); err != nil {
		panic(err)
	}
	strs := dat["height"].(float64)

	object, err := json.Marshal(strs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(object)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
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

func zerotonull(n float64) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%v", n)
}

func fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, err
}

func preprocess(c context.Context) {
	var d StoredDataSet
	if err1 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err1 != nil && err1 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Bitcoin data: %v", err1)
		return
	}
	var dd StoredDataSetPrice
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSetPrice", "data", 0, nil), &dd); err != nil && err != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Bitcoin data: %v", err)
		return
	}
	var e StoredDataSet
	if err2 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "ether", 0, nil), &e); err2 != nil && err2 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Ether data: %v", err2)
		return
	}
	var l StoredDataSet
	if err3 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "ltc", 0, nil), &l); err3 != nil && err3 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching LTC data: %v", err3)
		return
	}
	var gold StoredDataSet
	if err4 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "GOLDAMGBD228NLBM", 0, nil), &gold); err4 != nil && err4 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Gold data: %v", err4)
		return
	}
	var euro StoredDataSet
	if err5 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSEU", 0, nil), &euro); err5 != nil && err5 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Euro data: %v", err5)
		return
	}
	var pound StoredDataSet
	if err6 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSUK", 0, nil), &pound); err6 != nil && err6 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching GBP data: %v", err6)
		return
	}
	var brazil StoredDataSet
	if err7 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXBZUS", 0, nil), &brazil); err7 != nil && err7 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching BRL data: %v", err7)
		return
	}
	var china StoredDataSet
	if err8 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXCHUS", 0, nil), &china); err8 != nil && err8 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching China data: %v", err8)
		return
	}
	var thailand StoredDataSet
	if err9 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXTHUS", 0, nil), &thailand); err9 != nil && err9 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Thailand data: %v", err9)
		return
	}
	var japan StoredDataSet
	if err10 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXJPUS", 0, nil), &japan); err10 != nil && err10 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching Japan data: %v", err10)
		return
	}
	var southafrica StoredDataSet
	if err11 := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXSFUS", 0, nil), &southafrica); err11 != nil && err11 != datastore.ErrNoSuchEntity {
		log.Printf("Fetching South Africa data: %v", err11)
		return
	}

	bitcoindata = d.Data[29:]
	bitcoindataprice = dd.Data[29 : len(dd.Data)-1]
	bitcoinEther = e.Data[29:]
	bitcoinLTC = l.Data[29:]
	bitcoinGOLDAMGBD228NLBM = gold.Data[29:]
	bitcoinDEXUSEU = euro.Data[29:]
	bitcoinDEXUSUK = pound.Data[29:]
	bitcoinDEXBZUS = brazil.Data[29:]
	bitcoinDEXCHUS = china.Data[29:]
	bitcoinDEXTHUS = thailand.Data[29:]
	bitcoinDEXJPUS = japan.Data[29:]
	bitcoinDEXSFUS = southafrica.Data[29:]

	latest = d.Data[len(d.Data)-1]
	latestPrice = dd.Data[len(dd.Data)-1].Price
	renderdata.Data = stripdata(bitcoindata)
	renderdata.Ether = stripdata(bitcoinEther)
	renderdata.GOLDAMGBD228NLBM = stripdata(bitcoinGOLDAMGBD228NLBM)
	renderdata.DEXUSEU = stripdata(bitcoinDEXUSEU)
	renderdata.DEXUSUK = stripdata(bitcoinDEXUSUK)
	renderdata.DEXBZUS = stripdata(bitcoinDEXBZUS)
	renderdata.DEXCHUS = stripdata(bitcoinDEXCHUS)
	renderdata.DEXTHUS = stripdata(bitcoinDEXTHUS)
	renderdata.DEXJPUS = stripdata(bitcoinDEXJPUS)
	renderdata.DEXSFUS = stripdata(bitcoinDEXSFUS)
	renderdata.Latest30 = d.Data[len(d.Data)-1].Volatility
	renderdata.Latest60 = d.Data[len(d.Data)-1].Volatility60
}

func stripdata(data DataSet) BasicDataSet {
	return BasicDataSet(data)
}

func (s DataSet) Len() int {
	return len(s)
}

func (s DataSet) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DataSet) Less(i, j int) bool {
	return s[i].Date < s[j].Date
}

func (s DataSetPrice) Len() int {
	return len(s)
}

func (s DataSetPrice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s DataSetPrice) Less(i, j int) bool {
	return s[i].Date < s[j].Date
}
