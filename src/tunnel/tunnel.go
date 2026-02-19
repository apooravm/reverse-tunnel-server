package tunnel

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	TManager = NewTunnelManager()
)

type Tunnel struct {
	Conn *websocket.Conn
	Mu   sync.Mutex
}

type TunnelManager struct {
	Mu      sync.RWMutex
	Tunnels map[string]*Tunnel
}

func NewTunnelManager() *TunnelManager {
	return &TunnelManager{
		Tunnels: make(map[string]*Tunnel),
	}
}

func (tm *TunnelManager) Register(sub string, t *Tunnel) {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()
	tm.Tunnels[sub] = t
}

func (tm *TunnelManager) Get(sub string) (*Tunnel, bool) {
	tm.Mu.RLock()
	defer tm.Mu.RUnlock()
	t, ok := tm.Tunnels[sub]
	return t, ok
}

func (tm *TunnelManager) Remove(sub string) {
	tm.Mu.Lock()
	defer tm.Mu.Unlock()
	delete(tm.Tunnels, sub)
}
