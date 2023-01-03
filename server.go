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
		log.Print("upgrade:", err)
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

func healthCheckPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "OK")
}

func serverStatsPage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, monitoring.GetMonitor().GetStats())
}

// flags
var addr = flag.String("addr", "localhost:8080", "http service address")
var metricsEnabled = flag.Bool("metricsEnabled", true, "flag to enable metrics")

func main() {
	fmt.Println("Hello, this is the POC game server.")
	flag.Parse()

	log.SetFlags(0)
	config.LoadConfig(&config.Config{
		Addr:           *addr,
		MetricsEnabled: *metricsEnabled,
	})

	http.HandleFunc("/echo", echo)
	http.HandleFunc("/health", healthCheckPage)
	http.HandleFunc("/stats", serverStatsPage)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
