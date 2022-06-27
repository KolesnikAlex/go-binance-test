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
	wg              sync.WaitGroup
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
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			doneC, _, err := binance.WsAggTradeServe(fullListTickers[i].Symbol, wsAggTradeHandler, errHandler)
			if err != nil {
				fmt.Println(err)
				return
			}
			<-doneC
		}(i)
	}

	for i := 0; ; i++ {
		time.Sleep(timeOut * time.Second)
		for key, val := range tickers {
			fmt.Println(key, val)
			delete(tickers, key)
		}
		fmt.Println()
	}

	wg.Wait()
}
