package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type BinanceApiJson struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
	Msg    string `json:"msg"`
}

func getCryptoAPI(cryptoTitle string) BinanceApiJson {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/ticker/price?symbol="+cryptoTitle+"USDT", nil)

	if err != nil {
		log.Fatal(err)
	}
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var ticker BinanceApiJson
	err = json.Unmarshal(body, &ticker)

	if err != nil {
		log.Fatal(err)
	}

	return ticker
}
