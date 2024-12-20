#!/usr/bin/bash

if [ "$#" -ne 1 ]; then
    echo "No input file given"
    echo "Usage: ./run.sh <path_to_xls_file>"
    exit 1
fi
go build -o bin/summarizer.exe .\\cmd\\tradesummarize\\ && .\\bin\\summarizer.exe $1