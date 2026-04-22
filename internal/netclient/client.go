package netclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	gamemath "ascii-game/internal/math"
	"ascii-game/internal/proto"
	"ascii-game/internal/sim"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn          *websocket.Conn
	writeMu       sync.Mutex
	localPlayerID sim.PlayerID

	mu       sync.RWMutex
	snapshot sim.Snapshot

	errMu   sync.Mutex
	err     error
	closeMu sync.Once
}

func Connect(addr string) (*Client, error) {
	endpoint, err := websocketURL(addr)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{
		HandshakeTimeout: 5 * time.Second,
	}

	conn, resp, err := dialer.Dial(endpoint.String(), http.Header{})
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("websocket dial failed with status %s: %w", resp.Status, err)
		}
		return nil, err
	}

	if err := conn.WriteMessage(websocket.BinaryMessage, proto.EncodeMessage(proto.Message{
		Type:    proto.MsgJoin,
		Payload: proto.MarshalJoin(),
	})); err != nil {
		_ = conn.Close()
		return nil, err
	}

	_, payload, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	msg, err := proto.DecodeMessageBytes(payload)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if msg.Type != proto.MsgJoinAck {
		_ = conn.Close()
		return nil, fmt.Errorf("expected join ack, got message type 0x%02x", msg.Type)
	}

	playerID, err := proto.UnmarshalJoinAck(msg.Payload)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	client := &Client{
		conn:          conn,
		localPlayerID: playerID,
	}

	go client.readLoop()

	return client, nil
}

func websocketURL(addr string) (url.URL, error) {
	if strings.HasPrefix(addr, "ws://") || strings.HasPrefix(addr, "wss://") {
		parsed, err := url.Parse(addr)
		if err != nil {
			return url.URL{}, err
		}
		if parsed.Path == "" {
			parsed.Path = "/ws"
		}
		return *parsed, nil
	}

	scheme := "wss"
	if isLocalAddress(addr) {
		scheme = "ws"
	}

	return url.URL{
		Scheme: scheme,
		Host:   addr,
		Path:   "/ws",
	}, nil
}

func isLocalAddress(addr string) bool {
	host := addr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host = addr[:idx]
	}

	return host == "127.0.0.1" || host == "localhost" || strings.HasPrefix(host, "0.0.0.0")
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

	c.writeMu.Lock()
	err = c.conn.WriteMessage(websocket.BinaryMessage, proto.EncodeMessage(proto.Message{
		Type:    proto.MsgMove,
		Payload: payload,
	}))
	c.writeMu.Unlock()
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
		_, payload, err := c.conn.ReadMessage()
		if err != nil {
			c.setError(err)
			return
		}

		msg, err := proto.DecodeMessageBytes(payload)
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
