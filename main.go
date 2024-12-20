package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/extrame/xls"
)

type Trade struct {
	Ticker string
	Type   string
	Amount float64
	Isin   string
	Shares int
	Date   string
}

type Trades struct {
	Ticker                 string
	SharesToCount          int
	SharesToCountForBuying int
	Transactions           []Trade
	Buy                    []Trade
	Sell                   []Trade
}

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No input file provided")
		Summarize("data/ND_JAN_FEB.xls")
		return
	}
	fmt.Println(args)
	Summarize(args[0])
}

func Summarize(path string) {
	f1, err := xls.Open(path, "utf-8")
	if err != nil {
		log.Fatalf("Cannot open file: %v", err)
	}

	allTrades, err := parseTrades(f1)
	if err != nil {
		log.Fatalf("Error parsing trades: %v", err)
	}

	trades := organizeTrades(allTrades)

	printSummary(trades)
}

func parseTrades(f1 *xls.WorkBook) ([]Trade, error) {
	allTrades := make([]Trade, 0)
	sheet := f1.GetSheet(0)

	if sheet == nil {
		return nil, fmt.Errorf("cannot get sheet: %v", sheet)
	}

	for i := 1; i <= int(sheet.MaxRow); i++ { // Skip header row
		row := sheet.Row(i)
		if row == nil || row.LastCol() < 1 {
			continue
		}

		t, err := parseTradeRow(row)
		if err != nil {
			fmt.Printf("Error parsing row %d: %v\n", i, err)
			continue
		}

		allTrades = append(allTrades, t)
	}

	return allTrades, nil
}

/* type Config struct {
	Isin   *int `json:"isin"`
	Type   *int `json:"type"`
	Ticker *int `json:"ticker"`
	Date   *int `json:"date"`
	Shares *int `json:"shares"`
	Amount *int `json:"amount"`
} */

func parseTradeRow(row *xls.Row) (Trade, error) {
	var t Trade
	var err error

	t.Isin = row.Col(0)
	t.Type = strings.TrimSpace(row.Col(1))
	t.Ticker = row.Col(2)
	t.Date = row.Col(6)

	t.Shares, err = strconv.Atoi(row.Col(8))
	if err != nil {
		return t, fmt.Errorf("invalid share count: %v "+t.Ticker, err)
	}

	t.Amount, err = strconv.ParseFloat(row.Col(22), 64)
	if err != nil {
		return t, fmt.Errorf("invalid amount: %v", err)
	}

	return t, nil
}

func organizeTrades(allTrades []Trade) map[string]Trades {
	trades := make(map[string]Trades)

	for _, t := range allTrades {
		value, exists := trades[t.Ticker]
		if !exists {
			value = Trades{Ticker: t.Ticker}
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

func printSummary(trades map[string]Trades) {
	total := 0.0
	totalWins := 0
	totalLosses := 0

	for _, vTrades := range trades {
		if len(vTrades.Buy) == 0 || len(vTrades.Sell) == 0 {
			continue
		}

		if vTrades.SharesToCountForBuying < vTrades.SharesToCount {
			vTrades.SharesToCount = vTrades.SharesToCountForBuying
		}

		totalBuys, totalSells := calculatePnL(vTrades)
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
}

func calculatePnL(vTrades Trades) (totalBuys, totalSells float64) {
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
				vTrades.SharesToCount = 0
				break
			}
			totalSells += trd.Amount
			vTrades.SharesToCount -= trd.Shares
			buyAmount -= trd.Shares
		case "Osto":
			if !hasAnySellsLeft(vTrades.Transactions[:i]) {
				continue
			}
			totalBuys += trd.Amount
			buyAmount += trd.Shares
			shouldCountSells = true
		}
	}

	return totalBuys, totalSells
}

func hasAnySellsLeft(transactions []Trade) bool {
	for i := len(transactions) - 1; i >= 0; i-- {
		if transactions[i].Type == "Myynti" {
			return true
		}
	}
	return false
}
