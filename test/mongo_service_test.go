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
var test_collection *mongo.Collection
var ctx context.Context
var test_data tradesummarize.ApiTrades

func TestMain(m *testing.M) {
	var err error
	client, ctx, err = tradesummarize.ConnectMongoDB("mongodb://localhost:27017")
	if err != nil {
		log.Fatal(err)
	}

	expectedTrade := tradesummarize.Trade{Id: "01234567", Ticker: "TestTicker", Type: "Osto", Amount: 123.123, Isin: "ISINHERE", Shares: 12, Date: "11.11.2011"}
	transactions := append(make([]tradesummarize.Trade, 0), expectedTrade)
	test_data = tradesummarize.ApiTrades{Ticker: "TestTicker", Transactions: transactions}
	test_collection = client.Database("TradeDb-Collect_test").Collection("Trades")
	code := m.Run()

	client.Database("TradeDb-Collect_test").Drop(context.Background())
	client.Disconnect(ctx)
	os.Exit(code)
}

func TestSaveData(t *testing.T) {

	err := tradesummarize.SaveData(test_collection, test_data)
	if err != nil {
		t.Fatal("Error saving data to db")
	}

	var apiTrades tradesummarize.ApiTrades
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	err = test_collection.FindOne(ctx, bson.M{"Ticker": "TestTicker"}).Decode(&apiTrades)
	if err != nil {
		t.Fatal("Error Finding ticker", err.Error())
	}

	if len(apiTrades.Transactions) != 1 {
		t.Fatal("Error Wrong number of Transactions found: ", len(apiTrades.Transactions))
	}
	expectedTrade := test_data.Transactions[0]
	if expectedTrade != apiTrades.Transactions[0] {
		t.Fatal("Error Unexpected Trade found: ", apiTrades.Transactions[0], "Expected: ", expectedTrade)
	}
}

func TestFindByTransactionsByTicker(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	_, err := test_collection.InsertOne(ctx, test_data)
	if err != nil {
		t.Fatal("Error inserting data to db: ", err.Error())
	}

	trades := tradesummarize.FindByTransactionsByTicker(test_collection, test_data, test_data.Ticker)
	if trades == nil {
		t.Fatal("Error finding ticker", err.Error())
	}
	expectedTrade := test_data.Transactions[0]

	if !slices.Contains(*trades, expectedTrade) {
		t.Fatal("Error Unexpected transaction found in : ", *trades, "Expected:", expectedTrade)

	}
}

func TestUpdateTransactionForTicker(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	_, err := test_collection.InsertOne(ctx, test_data)
	if err != nil {
		t.Fatal("Error inserting data to db: ", err.Error())
	}
	newTrade := tradesummarize.Trade{Id: "01234567", Ticker: "TestTicker", Type: "Osto", Amount: 456.456, Isin: "ISINHERE", Shares: 12, Date: "11.11.2011"}

	err = tradesummarize.InsertNewTransactionForTicker(test_collection, newTrade.Ticker, newTrade)
	if err != nil {
		t.Fatal("Error updatuing data to db: ", err.Error())
	}

	var apiTrades tradesummarize.ApiTrades
	err = test_collection.FindOne(ctx, bson.M{"Ticker": newTrade.Ticker}).Decode(&apiTrades)
	if err != nil {
		t.Fatal("Error Finding ticker", err.Error())
	}

	if len(apiTrades.Transactions) != 2 {
		t.Fatal("Error Wrong number of Transactions found: ", len(apiTrades.Transactions))
	}

	expectedTrade := test_data.Transactions[0]

	if !slices.Contains(apiTrades.Transactions, expectedTrade) {
		t.Fatal("No Old trade found")
	}
	if !slices.Contains(apiTrades.Transactions, newTrade) {
		t.Fatal("No new trade Inserted")
	}
}
