## Usage

1. Download trade data from nordea investor
1. Execute app `summarizer path/to/excelfile`

## Operation

1. Parse input data based on config
2. Calculate PnL for tickers
3. Serve all trades using api

## TODO

1. Improve api serving
1. Add tests
1. Different api methods for querying different tickers and timeframes
1. Add language configuration
1. Improve transactions, remove unnecessary fields
1. Create different app timeline with Date and amount of sold stock and name of stock

- `Date` `Amount` `Ticker`
