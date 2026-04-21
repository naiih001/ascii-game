package server

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"

	gamemath "ascii-game/internal/math"
	"ascii-game/internal/proto"
	"ascii-game/internal/sim"
	"ascii-game/internal/world"
)

const ticksPerSecond = 30

type Server struct {
	addr       string
	listener   net.Listener
	sim        *sim.Simulation
	register   chan net.Conn
	unregister chan *clientSession
	inputs     chan playerInput
	clients    map[sim.PlayerID]*clientSession
}

type clientSession struct {
	conn     net.Conn
	playerID sim.PlayerID
	send     chan []byte
	once     sync.Once
}

type playerInput struct {
	playerID sim.PlayerID
	delta    gamemath.Vec2
}

func New(addr string) *Server {
	return &Server{
		addr:       addr,
		sim:        sim.New(world.NewDefaultMap()),
		register:   make(chan net.Conn),
		unregister: make(chan *clientSession, 32),
		inputs:     make(chan playerInput, 128),
		clients:    make(map[sim.PlayerID]*clientSession),
	}
}

func (s *Server) Run() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	s.listener = listener
	log.Printf("server listening on %s", listener.Addr())

	go s.acceptLoop()

	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	for {
		select {
		case conn := <-s.register:
			if err := s.addClient(conn); err != nil {
				log.Printf("register client: %v", err)
				_ = conn.Close()
			}
		case client := <-s.unregister:
			s.removeClient(client)
		case input := <-s.inputs:
			s.sim.QueueMove(input.playerID, input.delta)
		case <-ticker.C:
			s.sim.Tick()
			s.broadcastSnapshot()
		}
	}
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("accept client: %v", err)
			continue
		}

		s.register <- conn
	}
}

func (s *Server) addClient(conn net.Conn) error {
	msg, err := proto.DecodeMessage(conn)
	if err != nil {
		return err
	}

	if msg.Type != proto.MsgJoin {
		return errors.New("first message must be join")
	}

	if err := proto.UnmarshalJoin(msg.Payload); err != nil {
		return err
	}

	playerID := s.sim.AddPlayer(spawnPoint(len(s.clients)))
	client := &clientSession{
		conn:     conn,
		playerID: playerID,
		send:     make(chan []byte, 32),
	}
	s.clients[playerID] = client

	client.send <- proto.EncodeMessage(proto.Message{
		Type:    proto.MsgJoinAck,
		Payload: proto.MarshalJoinAck(playerID),
	})
	client.send <- proto.EncodeMessage(proto.Message{
		Type:    proto.MsgSnapshot,
		Payload: proto.MarshalSnapshot(s.sim.Snapshot()),
	})

	go s.readLoop(client)
	go s.writeLoop(client)
	return nil
}

func (s *Server) removeClient(client *clientSession) {
	if client == nil {
		return
	}

	current, ok := s.clients[client.playerID]
	if !ok || current != client {
		return
	}

	delete(s.clients, client.playerID)
	s.sim.RemovePlayer(client.playerID)
	client.close()
}

func (s *Server) broadcastSnapshot() {
	encoded := proto.EncodeMessage(proto.Message{
		Type:    proto.MsgSnapshot,
		Payload: proto.MarshalSnapshot(s.sim.Snapshot()),
	})

	for _, client := range s.clients {
		select {
		case client.send <- encoded:
		default:
			log.Printf("dropping slow client %d", client.playerID)
			s.removeClient(client)
		}
	}
}

func (s *Server) readLoop(client *clientSession) {
	for {
		msg, err := proto.DecodeMessage(client.conn)
		if err != nil {
			s.unregister <- client
			return
		}

		switch msg.Type {
		case proto.MsgMove:
			dx, dy, err := proto.UnmarshalMove(msg.Payload)
			if err != nil {
				s.unregister <- client
				return
			}

			s.inputs <- playerInput{
				playerID: client.playerID,
				delta:    gamemath.Vec2{X: dx, Y: dy},
			}
		default:
			s.unregister <- client
			return
		}
	}
}

func (s *Server) writeLoop(client *clientSession) {
	for msg := range client.send {
		if _, err := client.conn.Write(msg); err != nil {
			s.unregister <- client
			return
		}
	}
}

func (c *clientSession) close() {
	c.once.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

func spawnPoint(index int) gamemath.Vec2 {
	spawnOffsets := []gamemath.Vec2{
		{X: 2, Y: 2},
		{X: 4, Y: 2},
		{X: 6, Y: 2},
		{X: 8, Y: 2},
	}

	if index < len(spawnOffsets) {
		return spawnOffsets[index]
	}

	return gamemath.Vec2{
		X: 2 + ((index % 6) * 2),
		Y: 4 + ((index / 6) * 2),
	}
}
