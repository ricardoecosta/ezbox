package main

// todo: use errors package everywhere

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"time"

	"fmt"

	"github.com/gorilla/websocket"
	"github.com/ricardoecosta/ezbox/gpio"
	"github.com/ricardoecosta/ezbox/media"
	"github.com/ricardoecosta/ezbox/session"
)

var (
	upgrader   = websocket.Upgrader{}
	testTicker = time.NewTicker(1 * time.Second)
)

func main() {
	config := LoadConfig("conf.json")
	go waitForGracefulShutdownSignal()

	media.InitPlayer("") // todo: read from config
	media.InitCollection(config.MediaDirectories)
	media.PrintIndexedMediaCollection()

	pins := FlattenedPins(config.Controls)
	stream := gpio.Stream(pins)

	go func() {
		for pin := range stream {

			// todo
			m := NewChannelChangedMessage(
				"channel "+strconv.Itoa(int(pin.Number)),
				strconv.Itoa(int(pin.Value)),
				"")

			log.Printf("pin update, number=%v value=%v", pin.Number, pin.Value)
			session.BroadcastMessageToSessions(m)
		}
	}()

	// frontend file server
	http.Handle("/", http.FileServer(http.Dir(config.FrontendRoot)))

	// websocket endpoint
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println("error upgrading websockets", err)
			return
		}

		session.NewSession(request.Header.Get("sec-websocket-key"), conn)
	})

	log.Printf("http server listening, port=%v", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}

func waitForGracefulShutdownSignal() {
	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGTERM)
	signal.Notify(sigs, syscall.SIGINT)

	shutdownSignal := <-sigs
	log.Printf("caught %v signal, shutting down gracefully", shutdownSignal)

	testTicker.Stop()
	session.TerminateSessions()
	os.Exit(0)
}
