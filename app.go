package main

import (
	"github.com/pion/webrtc/v3"
	"sync"
)

type App struct {
	mu      sync.RWMutex
	rooms   map[string]*Room
	viewers map[string]*PeerConnectionEgress
}

func NewApp() *App {
	return &App{
		mu:      sync.RWMutex{},
		rooms:   make(map[string]*Room),
		viewers: make(map[string]*PeerConnectionEgress),
	}
}

func (a *App) AddViewer(offer webrtc.SessionDescription) (*PeerConnectionEgress, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	egress, err := NewPeerConnectionEgress(offer)
	if err != nil {
		return nil, err
	}

	egress.peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		if state == webrtc.PeerConnectionStateClosed {
			a.Switch(egress.ID, "")
		}
	})

	a.viewers[egress.ID] = egress
	return egress, nil
}

func (a *App) AddRoom(port int) (*Room, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	r, err := NewRoom(port)
	if err != nil {
		return nil, err
	}
	a.rooms[r.ID] = r
	return r, nil
}

func (a *App) Switch(viewerID string, targetRoomID string) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	targetRoom, hasTargetRoom := a.rooms[targetRoomID]

	v, ok := a.viewers[viewerID]
	if ok {
		for _, room := range a.rooms {
			if room.Has(viewerID) {
				room.RemoveParticipant(v)
			}
		}
	}

	if hasTargetRoom {
		targetRoom.AddParticipant(v)
	}
}

func (a *App) ListRooms() []*Room {
	a.mu.RLock()
	defer a.mu.RUnlock()

	res := make([]*Room, 0, len(a.rooms))
	for _, room := range a.rooms {
		res = append(res, room)
	}

	return res
}
