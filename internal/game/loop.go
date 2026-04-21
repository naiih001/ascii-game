package game

import "time"

const ticksPerSecond = 30

func (g *Game) Run() error {
	defer g.Close()

	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	g.Render()

	for g.running {
		<-ticker.C
		g.drainInput()
		g.Update()
		g.Render()
	}

	return g.Err()
}

func (g *Game) drainInput() {
	for {
		select {
		case action := <-g.input:
			g.handleAction(action)
		default:
			return
		}
	}
}
