package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/extrame/xls"

	"github.com/jsipola/TradeSummarizer/internal/app"
	"github.com/jsipola/TradeSummarizer/internal/helpers"
)

var tradesData map[string][]app.Trade
var tradesData2 []app.ApiTrades

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No input file provided")
		Summarize("data/ND_JAN_FEB.xls")
		return
	}
	tradesData = Summarize(args[0])

	http.HandleFunc("/api/trades", tradesHandler)
	http.HandleFunc("/api/validTrades", tradesHandler2)

	fmt.Println("Server started at http://localhost:8080")
	app.MongoInit(tradesData2)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func tradesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(tradesData)
}

func tradesHandler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	json.NewEncoder(w).Encode(tradesData2)
}

func Summarize(path string) map[string][]app.Trade {
	f1, err := xls.Open(path, "utf-8")
	if err != nil {
		log.Fatalf("Cannot open file: %v", err)
	}

	allTrades, err := parseTrades(f1)
	if err != nil {
		log.Fatalf("Error parsing trades: %v", err)
	}

	trades := organizeTrades(allTrades)

	executedTrades := printSummary(trades)
	return executedTrades
}

func parseTrades(f1 *xls.WorkBook) ([]app.Trade, error) {
	allTrades := make([]app.Trade, 0)
	sheet := f1.GetSheet(0)

	if sheet == nil {
		return nil, fmt.Errorf("cannot get sheet: %v", sheet)
	}

	cfg := helpers.ReadJsonConfig("default.json")

	for i := 1; i <= int(sheet.MaxRow); i++ { // Skip header row
		row := sheet.Row(i)
		if row == nil || row.LastCol() < 1 {
			continue
		}

		t, err := parseTradeRow(row, cfg)
		if err != nil {
			fmt.Printf("Error parsing row %d: %v\n", i, err)
			continue
		}

		allTrades = append(allTrades, t)
	}

	return allTrades, nil
}

func parseTradeRow(row *xls.Row, cfg helpers.Config) (app.Trade, error) {
	var t app.Trade
	var err error

	t.Id = strings.TrimSpace(row.Col(cfg.Id))
	t.Isin = strings.TrimSpace(row.Col(cfg.Isin))
	var typeTemp = strings.TrimSpace(row.Col(cfg.Type))
	t.Type = parseType(typeTemp)
	t.Ticker = strings.TrimSpace(row.Col(cfg.Ticker))
	t.Date = row.Col(cfg.Date)

	t.Shares, err = strconv.Atoi(row.Col(cfg.Shares))
	if err != nil {
		return t, fmt.Errorf("invalid share count: %v", err)
	}

	t.Amount, err = strconv.ParseFloat(row.Col(cfg.Amount), 64)
	if err != nil {
		return t, fmt.Errorf("invalid amount: %v", err)
	}

	return t, nil
}

func parseType(str string) string {
	switch str {
	case "Buy":
		return "Osto"
	case "Sell":
		return "Myynti"
	default:
		return str
	}
}

func organizeTrades(allTrades []app.Trade) map[string]app.Trades {
	trades := make(map[string]app.Trades)

	for _, t := range allTrades {
		value, exists := trades[t.Ticker]
		if !exists {
			value = app.Trades{Ticker: t.Ticker}
		}

		switch t.Type {
		case "Myynti":
			value.Sell = append(value.Sell, t)
			value.SharesToCount += t.Shares
		case "Osto":
			value.Buy = append(value.Buy, t)
			value.SharesToCountForBuying += t.Shares
		}

		value.Transactions = append(value.Transactions, t)
		trades[t.Ticker] = value
	}

	return trades
}

func printSummary(trades map[string]app.Trades) map[string][]app.Trade {
	total := 0.0
	totalWins := 0
	totalLosses := 0
	executedTrades := make(map[string]app.Trades)
	apiTrades := map[string][]app.Trade{}
	for _, vTrades := range trades {
		if len(vTrades.Buy) == 0 || len(vTrades.Sell) == 0 {
			continue
		}

		if vTrades.SharesToCountForBuying < vTrades.SharesToCount {
			vTrades.SharesToCount = vTrades.SharesToCountForBuying
		}

		executedTrades[vTrades.Ticker] = vTrades
		totalBuys := 0.0
		totalSells := 0.0
		totalBuys, totalSells, validTrades := calculatePnL(vTrades)
		apiTrades[vTrades.Ticker] = validTrades
		tradesData2 = append(tradesData2, app.ApiTrades{Ticker: vTrades.Ticker, Transactions: validTrades})
		total += totalSells - totalBuys

		amount := strconv.FormatFloat(totalSells-totalBuys, 'f', 2, 64)
		if totalSells-totalBuys > 0 {
			totalWins++
		} else {
			totalLosses++
		}

		fmt.Printf("Ticker: %s\n  ----> PnL: %s\n", vTrades.Ticker, amount)
	}

	fmt.Printf("Total: %.2f\n", total)
	fmt.Printf("Wins: %d Losses: %d\n", totalWins, totalLosses)
	return apiTrades
}

func calculatePnL(vTrades app.Trades) (totalBuys float64, totalSells float64, validTrades []app.Trade) {
	shouldCountSells := false
	buyAmount := 0
	for i := len(vTrades.Transactions) - 1; i >= 0; i-- {
		trd := vTrades.Transactions[i]
		if vTrades.SharesToCount == 0 {
			break
		}

		switch trd.Type {
		case "Myynti":
			if !shouldCountSells || buyAmount == 0 {
				continue
			}
			if vTrades.SharesToCount < trd.Shares || buyAmount < trd.Shares {
				totalSells += (trd.Amount / float64(trd.Shares)) * float64(vTrades.SharesToCount)
				partialTrd := trd
				partialTrd.Amount = (trd.Amount / float64(trd.Shares)) * float64(vTrades.SharesToCount)
				partialTrd.Shares = vTrades.SharesToCount
				vTrades.SharesToCount = 0
				validTrades = append(validTrades, partialTrd)
				break
			}
			validTrades = append(validTrades, trd)
			totalSells += trd.Amount
			vTrades.SharesToCount -= trd.Shares
			buyAmount -= trd.Shares
		case "Osto":
			if !hasAnySellsLeft(vTrades.Transactions[:i]) {
				continue
			}
			validTrades = append(validTrades, trd)
			totalBuys += trd.Amount
			buyAmount += trd.Shares
			shouldCountSells = true
		default:
			fmt.Println("No case found")
		}
	}

	return totalBuys, totalSells, validTrades
}

func hasAnySellsLeft(transactions []app.Trade) bool {
	for i := len(transactions) - 1; i >= 0; i-- {
		if transactions[i].Type == "Myynti" {
			return true
		}
	}
	return false
}
