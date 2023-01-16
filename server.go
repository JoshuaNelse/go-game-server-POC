package main

import (
	"flag"
	"fmt"
	"game-poc/server/config"
	"game-poc/server/monitoring"
	"game-poc/server/network/client"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading http request to websocket:", err)
		return
	}
	//  Metric Instrumentation
	if config.GetConfig().MetricsEnabled {
		monitor := monitoring.GetMonitor()
		monitor.AddConnection(c)
		defer monitor.RemoveConnection(c)
	}

	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

var clients []*client.Client

func game(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading http request to websocket:", err)
		return
	}
	// need to eventually find a way to lock or prevent collision here for client id
	clientId := uuid.New()
	gameClient := client.NewClient(clientId, c, gameChan)
	defer func() {
		defer c.Close()
		// TODO clean up client from clients list
	}()
	clients = append(clients, gameClient)

	gameClient.Listen()
}

func healthCheckPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func serverStatsPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, monitoring.GetMonitor().GetStats())
}

// flags
var addr = flag.String("addr", "localhost:8080", "http service address")
var metricsEnabled = flag.Bool("metricsEnabled", true, "flag to enable metrics")

// test global channel
var gameChan chan *[]byte

func main() {
	fmt.Println("Hello, this is the POC game server.")
	flag.Parse()

	log.SetFlags(0)
	config.LoadConfig(&config.Config{
		Addr:           *addr,
		MetricsEnabled: *metricsEnabled,
	})
	gameChan = make(chan *[]byte)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/game", game)
	http.HandleFunc("/health", healthCheckPage)
	http.HandleFunc("/stats", serverStatsPage)
	log.Fatal(http.ListenAndServe(*addr, nil))

}
