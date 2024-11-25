package test

import (
	"context"
	"log"
	"os"
	"slices"
	"testing"
	"time"

	tradesummarize "github.com/jsipola/TradeSummarizer/internal/app"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var client *mongo.Client
var ctx context.Context

func TestMain(m *testing.M) {
	var err error
	client, ctx, err = tradesummarize.ConnectMongoDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	client.Database("TradeDb-Collect_test").Drop(context.Background())
	client.Disconnect(ctx)
	os.Exit(code)
}

func TestSaveData(t *testing.T) {
	transactions := make([]tradesummarize.Trade, 0)
	expectedTrade := tradesummarize.Trade{Id: "01234567", Ticker: "TestTicker", Type: "Osto", Amount: 123.123, Isin: "ISINHERE", Shares: 12, Date: "11.11.2011"}
	transactions = append(transactions, expectedTrade)
	data := tradesummarize.ApiTrades{Ticker: "TestTicker", Transactions: transactions}

	db := client.Database("TradeDb-Collect_test")
	err := tradesummarize.SaveData(db, data)
	if err != nil {
		t.Fatal("Error saving data to db")
	}

	var apiTrades tradesummarize.ApiTrades
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	err = db.Collection("Trades").FindOne(ctx, bson.M{"Ticker": "TestTicker"}).Decode(&apiTrades)
	if err != nil {
		t.Fatal("Error Finding ticker", err.Error())
	}

	if len(apiTrades.Transactions) != 1 {
		t.Fatal("Error Wrong number of Transactions found: ", len(apiTrades.Transactions))
	}

	if expectedTrade != apiTrades.Transactions[0] {
		t.Fatal("Error Unexpected Trade found: ", apiTrades.Transactions[0], "Expected: ", expectedTrade)
	}
}

func TestFindByTransactionsByTicker(t *testing.T) {
	//TODO refactor setup method
	transactions := make([]tradesummarize.Trade, 0)
	expectedTrade := tradesummarize.Trade{Id: "01234567", Ticker: "TestTicker", Type: "Osto", Amount: 123.123, Isin: "ISINHERE", Shares: 12, Date: "11.11.2011"}
	transactions = append(transactions, expectedTrade)
	data := tradesummarize.ApiTrades{Ticker: "TestTicker", Transactions: transactions}

	db := client.Database("TradeDb-Collect_test")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	_, err := db.Collection("Trades").InsertOne(ctx, data)
	if err != nil {
		t.Fatal("Error inserting data to db: ", err.Error())
	}

	trades := tradesummarize.FindByTransactionsByTicker(db, data, data.Ticker)
	if trades == nil {
		t.Fatal("Error finding ticker", err.Error())
	}
	if !slices.Contains(*trades, expectedTrade) {
		t.Fatal("Error Unexpected transaction found in : ", *trades, "Expected:", expectedTrade)

	}
}
