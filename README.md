## Preface

1. Download trade data from nordea investor

## Build & Execute

1. `go build -o bin/summarizer.exe .\cmd\tradesummarize\`
1. `.\bin\summarizer.exe /path/to/excelfile`

Generates a default config file and starts a server for data `Server started at http://localhost:8080/api/validTrades`

## Operation

1. Parse input data based on config
2. Calculate PnL for tickers
3. Serve all trades using api

## TODO

1. Improve api serving
1. Add more tests
1. Different api methods for querying different tickers and timeframes
1. Add language configuration
1. Improve transactions, remove unnecessary fields
1. Add functionality to save result to data
   - Add methods to update data
   - Add dividend collection methods
