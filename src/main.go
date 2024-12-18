package main

import (
	"bufio"
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	botapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/willmroliver/plathbot/src/io"
)

const (
	DonateLink string = "https://support.wwf.org.uk/"
	AdoptLink  string = "https://gifts.worldwildlife.org/gift-center/gifts/species-adoptions/duck-billed-platypus"
)

var (
	bot *botapi.BotAPI
)

func main() {
	// Create a custom HTTP transport with DialContext
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Prefer IPv4
			return (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext(ctx, "tcp4", addr)
		},
		// Optional: additional transport settings
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	// Create a custom HTTP client with the transport
	httpClient := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	var err error
	bot, err = botapi.NewBotAPI("7323800698:AAE2RcvU-g81Iz-nNbRsnglTWmbCuZJJzJA")
	if err != nil {
		log.Panic(err)
	}

	bot.Client = httpClient

	u := botapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	ch := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, ch)

	log.Println("Listening for updates. Press enter to exit")

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates botapi.UpdatesChannel) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			io.NewBotContext(bot, update).HandleUpdate()
		}
	}
}
