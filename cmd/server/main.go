package main

import (
	"flag"
	"log"
	"os"

	"ascii-game/internal/server"
)

func main() {
	addr := flag.String("addr", defaultListenAddr(), "server listen address")
	flag.Parse()

	srv := server.New(*addr)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

func defaultListenAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return "0.0.0.0:" + port
	}

	return "127.0.0.1:7777"
}
