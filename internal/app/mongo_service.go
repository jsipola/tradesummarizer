package app

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongoDB(uri string) (*mongo.Client, context.Context, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println("Connected to MongoDB!")
	return client, ctx, nil
}

/* func FindByTicker(client *mongo.Client, data ApiTrades) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	collection := client.Database("TradeDb-Collect").Collection("Trades")

	var trades ApiTrades
	err := collection.FindOne(ctx, bson.M{"Ticker": data.Ticker}).Decode(trades)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return nil
	}
	return &trades
} */

func FindByTransactionsByTicker(db *mongo.Database, data ApiTrades, ticker string) *[]Trade {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	collection := db.Collection("Trades")

	var trades ApiTrades
	err := collection.FindOne(ctx, bson.M{"Ticker": ticker}).Decode(&trades)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return nil
	}

	return &trades.Transactions
}

func SaveData(db *mongo.Database, data ApiTrades) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	collection := db.Collection("Trades")

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		return err
	}

	fmt.Println("Data saved successfully!")
	return nil
}

func UpdateTransactionForTicker(client *mongo.Client, ticker string, trade Trade) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	collection := client.Database("TradeDb-Collect").Collection("Trades")

	_, err := collection.UpdateOne(ctx, bson.M{"Ticker": ticker}, bson.M{"$push": bson.M{"Transactions": trade}})
	if err != nil {
		return err
	}
	fmt.Println("Updated existing Ticker:", ticker, " with new transaction id:", trade.Id)
	return nil
}

func MongoInit(tradeData []ApiTrades) {
	uri := "mongodb://localhost:27017/TradeDb-Collect"

	client, ctx, err := ConnectMongoDB(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	for _, value := range tradeData {

		/* 		if value.Ticker == "YOU" {
			fmt.Println("YOU")
			value.Transactions = append(value.Transactions, Trade{"12345678", "YOU", "Osto", 123.123, "ISINHERE", 12, "11.11.2011"})
		} */
		db := client.Database("TradeDb-Collect")

		var existingTransactions = FindByTransactionsByTicker(db, value, value.Ticker)
		if existingTransactions == nil {
			// Save new Ticker
			err = SaveData(db, value)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			for _, transaction := range value.Transactions {
				// Use ContainsFunc instead?
				if slices.Contains(*existingTransactions, transaction) {
					fmt.Println("Found existing transactions id for Ticker:", value.Ticker)
				} else {
					UpdateTransactionForTicker(client, transaction.Ticker, transaction)
				}
			}
		}
	}
}
