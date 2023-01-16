package main

import (
	"flag"
	"fmt"
	"game-poc/server/config"
	"game-poc/server/monitoring"
	"log"
	"net/http"

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

func serverListenWrite(c *websocket.Conn, ch chan *[]byte) {
	defer c.Close()

	for {
		message := <-ch
		err := c.WriteMessage(1, *message)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("game sent: %s", message)
	}
}

func serverListenRead(c *websocket.Conn) {
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("game server recv: %s", message)

		// echo for now, but this should send inputs somewhere to be calculated and forwarded to all clients
		gameChan <- &message
	}
}

func game(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Error upgrading http request to websocket:", err)
		return
	}

	go serverListenRead(c)
	serverListenWrite(c, gameChan)
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
