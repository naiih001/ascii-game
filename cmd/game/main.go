package main

import (
	"flag"
	"log"

	"ascii-game/internal/game"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:7777", "server address")
	flag.Parse()

	g, err := game.New(*addr)
	if err != nil {
		log.Fatal(err)
	}

	if err := g.Run(); err != nil {
		log.Fatal(err)
	}
}
