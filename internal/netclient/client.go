package netclient

import (
	"fmt"
	"net"
	"sync"
	"time"

	gamemath "ascii-game/internal/math"
	"ascii-game/internal/proto"
	"ascii-game/internal/sim"
)

type Client struct {
	conn          net.Conn
	localPlayerID sim.PlayerID

	mu       sync.RWMutex
	snapshot sim.Snapshot

	errMu   sync.Mutex
	err     error
	closeMu sync.Once
}

func Connect(addr string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return nil, err
	}

	if _, err := conn.Write(proto.EncodeMessage(proto.Message{
		Type:    proto.MsgJoin,
		Payload: proto.MarshalJoin(),
	})); err != nil {
		conn.Close()
		return nil, err
	}

	msg, err := proto.DecodeMessage(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if msg.Type != proto.MsgJoinAck {
		conn.Close()
		return nil, fmt.Errorf("expected join ack, got message type 0x%02x", msg.Type)
	}

	playerID, err := proto.UnmarshalJoinAck(msg.Payload)
	if err != nil {
		conn.Close()
		return nil, err
	}

	client := &Client{
		conn:          conn,
		localPlayerID: playerID,
	}

	go client.readLoop()

	return client, nil
}

func (c *Client) LocalPlayerID() sim.PlayerID {
	return c.localPlayerID
}

func (c *Client) Snapshot() sim.Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot
}

func (c *Client) SendMove(delta gamemath.Vec2) error {
	payload, err := proto.MarshalMove(delta.X, delta.Y)
	if err != nil {
		return err
	}

	_, err = c.conn.Write(proto.EncodeMessage(proto.Message{
		Type:    proto.MsgMove,
		Payload: payload,
	}))
	if err != nil {
		c.setError(err)
	}
	return err
}

func (c *Client) Err() error {
	c.errMu.Lock()
	defer c.errMu.Unlock()
	return c.err
}

func (c *Client) Close() error {
	var err error
	c.closeMu.Do(func() {
		err = c.conn.Close()
	})
	return err
}

func (c *Client) readLoop() {
	for {
		msg, err := proto.DecodeMessage(c.conn)
		if err != nil {
			c.setError(err)
			return
		}

		switch msg.Type {
		case proto.MsgSnapshot:
			snapshot, err := proto.UnmarshalSnapshot(msg.Payload)
			if err != nil {
				c.setError(err)
				return
			}

			c.mu.Lock()
			c.snapshot = snapshot
			c.mu.Unlock()
		default:
			c.setError(fmt.Errorf("unexpected message type 0x%02x", msg.Type))
			return
		}
	}
}

func (c *Client) setError(err error) {
	if err == nil {
		return
	}

	c.errMu.Lock()
	if c.err == nil {
		c.err = err
	}
	c.errMu.Unlock()
	_ = c.Close()
}
