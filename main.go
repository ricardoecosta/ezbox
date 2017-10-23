package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/gorilla/websocket"
	"github.com/ricardoecosta/ezbox/config"
	"github.com/ricardoecosta/ezbox/gpio"
	"github.com/ricardoecosta/ezbox/session"
)

var (
	upgrader   = websocket.Upgrader{}
	testTicker = time.NewTicker(1 * time.Second)
)

func main() {
	serverConfig := config.Load("conf.json")
	go waitForGracefulShutdownSignal()

	stream := gpio.SubscribeToGpioStream(serverConfig.SimulatedGPIOEnabled)
	go func() {
		for pin := range stream {
			m := NewPinUpdatedMessage(pin)
			session.BroadcastMessageToSessions(m)
		}
	}()

	// serve frontend
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.ServeFile(writer, request, serverConfig.FrontendBootstrapPage)
	})

	// websocket connection establishment
	http.HandleFunc("/ws", func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Println("error upgrading websockets", err)
		}

		session.NewSession(request.Header.Get("sec-websocket-key"), conn)
	})
	log.Fatal(http.ListenAndServe(":8765", nil))
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
