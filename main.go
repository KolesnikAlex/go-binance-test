package main

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
)

type tickersParam struct {
	maxPrise       float64
	minPrise       float64
	numberOfTrades int
}

func (t tickersParam) String() string {
	return fmt.Sprintf(" \tnumber of trades: %-6d\t min price: %12f\t max price: %12f", t.numberOfTrades, t.minPrise, t.maxPrise)
}

var (
	apiKey                        = ""
	secretKey                     = ""
	numberOfTickers               = 20
	timeOut         time.Duration = 60
	mutex           sync.Mutex
)

func main() {

	client := binance.NewClient(apiKey, secretKey)


	tickers := make(map[string]tickersParam)

	wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
		currPrice, err := strconv.ParseFloat(event.Price, 64)
		if err != nil {
			fmt.Println(err)
			return
		}
		mutex.Lock()
		if val, ok := tickers[event.Symbol]; ok {
			if currPrice > tickers[event.Symbol].maxPrise {
				val.maxPrise = currPrice
			} else {
				val.minPrise = currPrice
			}
			val.numberOfTrades++
			tickers[event.Symbol] = val
		} else {
			tickers[event.Symbol] = tickersParam{
				maxPrise:       currPrice,
				minPrise:       currPrice,
				numberOfTrades: 1,
			}
		}
		mutex.Unlock()
	}

	errHandler := func(err error) {
		fmt.Println(err)
	}

	fullListTickers, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}


	for i := 0; i < numberOfTickers; i++ {
		go func(i int) {
			doneC, _, err := binance.WsAggTradeServe(fullListTickers[i].Symbol, wsAggTradeHandler, errHandler)
			if err != nil {
				fmt.Println(err)
				return
			}
			<-doneC
		}(i)
	}

	tt := time.NewTicker(timeOut * time.Second)
	fmt.Printf("wait %d seconds...\n", timeOut)
	for range tt.C {
		mutex.Lock()
		for key := range tickers {
			fmt.Println(key, tickers[key])
			delete(tickers, key)
		}
		mutex.Unlock()
		fmt.Println()
	}

}
