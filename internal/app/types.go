package app

type Trade struct {
	Id     string  `bson:"Id"`
	Ticker string  `bson:"Ticker"`
	Type   string  `bson:"Type"`
	Amount float64 `bson:"Amount"`
	Isin   string  `bson:"Isin"`
	Shares int     `bson:"Shares"`
	Date   string  `bson:"Date"`
}

type ApiTrades struct {
	Ticker       string  `bson:"Ticker"`
	Transactions []Trade `bson:"Transactions"`
}

type Trades struct {
	Ticker                 string
	SharesToCount          int
	SharesToCountForBuying int
	Transactions           []Trade
	Buy                    []Trade
	Sell                   []Trade
}
