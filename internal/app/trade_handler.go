package app

import (
	"encoding/json"
	"net/http"
)

var tradesData map[string][]Trade
var tradesData2 []ApiTrades

func TradesHandler(w http.ResponseWriter, r *http.Request) {

	setHeaders(w)
	json.NewEncoder(w).Encode(tradesData)
}

func ValidTradesHandler(w http.ResponseWriter, r *http.Request) {

	setHeaders(w)
	json.NewEncoder(w).Encode(tradesData2)
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func SetTradesData(input map[string][]Trade) {
	tradesData = input
}

func SetTradesData2(input []ApiTrades) {
	tradesData2 = input
}
