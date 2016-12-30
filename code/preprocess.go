package btcvolatility

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
)

type BasicDataPoint struct {
	Date         string
	Volatility   float64 `datastore:",noindex"`
	Volatility60 float64 `datastore:",noindex"`
}

type BasicDataSet []DataPoint

var delayedpreprocess = delay.Func("preprocess", preprocess)

func preprocess(c context.Context) {
	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		log.Infof(c, "Fetching Bitcoin data: %v", err)
	}
	var e StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "ether", 0, nil), &e); err != nil {
		log.Infof(c, "Fetching Ether data: %v", err)
	}
	var gold StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "GOLDAMGBD228NLBM", 0, nil), &gold); err != nil {
		log.Infof(c, "Fetching Gold data: %v", err)
	}
	var euro StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSEU", 0, nil), &euro); err != nil {
		log.Infof(c, "Fetching Euro data: %v", err)
	}
	var pound StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSUK", 0, nil), &pound); err != nil {
		log.Infof(c, "Fetching GBP data: %v", err)
	}
	var brazil StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXBZUS", 0, nil), &brazil); err != nil {
		log.Infof(c, "Fetching BRL data: %v", err)
	}
	var china StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXCHUS", 0, nil), &china); err != nil {
		log.Infof(c, "Fetching China data: %v", err)
	}
	var thailand StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXTHUS", 0, nil), &thailand); err != nil {
		log.Infof(c, "Fetching Thailand data: %v", err)
	}
	var japan StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXJPUS", 0, nil), &japan); err != nil {
		log.Infof(c, "Fetching Japan data: %v", err)
	}

	var southafrica StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXSFUS", 0, nil), &southafrica); err != nil {
		log.Infof(c, "Fetching South Africa data: %v", err)
	}

	bitcoindata = d.Data[29:]
	latest = d.Data[len(d.Data)-1]
	renderdata.Data = stripdata(bitcoindata)
	renderdata.Ether = stripdata(e.Data[29:])
	renderdata.GOLDAMGBD228NLBM = stripdata(gold.Data[29:])
	renderdata.DEXUSEU = stripdata(euro.Data[29:])
	renderdata.DEXUSUK = stripdata(pound.Data[29:])
	renderdata.DEXBZUS = stripdata(brazil.Data[29:])
	renderdata.DEXCHUS = stripdata(china.Data[29:])
	renderdata.DEXTHUS = stripdata(thailand.Data[29:])
	renderdata.DEXJPUS = stripdata(japan.Data[29:])
	renderdata.DEXSFUS = stripdata(southafrica.Data[29:])
	renderdata.Latest30 = d.Data[len(d.Data)-1].Volatility
	renderdata.Latest60 = d.Data[len(d.Data)-1].Volatility60
}

func stripdata(data DataSet) BasicDataSet {
	return BasicDataSet(data)
}
