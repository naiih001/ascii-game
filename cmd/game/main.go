package main

import (
	"flag"
	"log"

	"ascii-game/internal/game"
)

func main() {
	addr := flag.String("addr", "ascii-game-production.up.railway.app", "server address")
	flag.Parse()

	g, err := game.New(*addr)
	if err != nil {
		log.Fatal(err)
	}

	if err := g.Run(); err != nil {
		log.Fatal(err)
	}
}
