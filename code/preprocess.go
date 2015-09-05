package btcvolatility

import (
	"appengine"
	"appengine/datastore"
	"appengine/delay"
)

type BasicDataPoint struct {
	Date         string
	Volatility   float64 `datastore:",noindex"`
	Volatility60 float64 `datastore:",noindex"`
}

type BasicDataSet []DataPoint

var delayedpreprocess = delay.Func("preprocess", preprocess)

func preprocess(c appengine.Context) {
	c.Infof("start preprocess")

	var d StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "data", 0, nil), &d); err != nil {
		panic(err)
	}
	var gold StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "GOLDAMGBD228NLBM", 0, nil), &gold); err != nil {
		panic(err)
	}
	var euro StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXUSEU", 0, nil), &euro); err != nil {
		panic(err)
	}
	var brazil StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXBZUS", 0, nil), &brazil); err != nil {
		panic(err)
	}
	var china StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXCHUS", 0, nil), &china); err != nil {
		panic(err)
	}
	var thailand StoredDataSet
	if err := datastore.Get(c, datastore.NewKey(c, "StoredDataSet", "DEXTHUS", 0, nil), &thailand); err != nil {
		panic(err)
	}
	bitcoindata = d.Data[29:]
	latest = d.Data[len(d.Data)-1]
	renderdata.Data = stripdata(bitcoindata)
	renderdata.GOLDAMGBD228NLBM = stripdata(gold.Data[29:])
	renderdata.DEXUSEU = stripdata(euro.Data[29:])
	renderdata.DEXBZUS = stripdata(brazil.Data[29:])
	renderdata.DEXCHUS = stripdata(china.Data[29:])
	renderdata.DEXTHUS = stripdata(thailand.Data[29:])
	renderdata.Latest30 = d.Data[len(d.Data)-1].Volatility
	renderdata.Latest60 = d.Data[len(d.Data)-1].Volatility60
}

func stripdata(data DataSet) BasicDataSet {
	return BasicDataSet(data)
}
