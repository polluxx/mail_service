package collector

import (
	"fmt"
	"memdb"
	"time"
)

const (
	// Time in minutes after which message is expired
	messageExpirationDurution = 60
	// Time in seconds for ticker interval
	tickerInterval = 180
)

func Collect() {
	garbage := make(chan string)
	// create new ticker for garbage collect
	ticker := time.NewTicker(time.Second * tickerInterval)
	go Tick(ticker, garbage)

	for {
		fmt.Println(<-garbage)
	}
}

func Tick(ticker *time.Ticker, garbage chan<- string) {
	for t := range ticker.C {
		Clear(t, garbage)
	}
}

func Clear(tick time.Time, garbage chan<- string) {
	memdb.GetInstance().ClearNotRelevantMessages(messageExpirationDurution * time.Minute)

	garbage <- fmt.Sprintf("Tick at %v", tick)
}
