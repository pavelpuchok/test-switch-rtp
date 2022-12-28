package main

import (
	"github.com/google/uuid"
	"github.com/pion/rtp"
	"log"
	"sync"
)

type Room struct {
	mu           sync.RWMutex
	ingress      *UDPIngress
	participants map[string]*PeerConnectionEgress
	ID           string

	lastTimeStamp uint32

	kfBuffer *Buffer
}

func NewRoom(port int) (*Room, error) {
	r := &Room{
		ID:           uuid.NewString(),
		mu:           sync.RWMutex{},
		participants: make(map[string]*PeerConnectionEgress),
		kfBuffer:     NewBuffer(100),
	}

	ingress, err := NewUDPIngress(r, port)
	if err != nil {
		return nil, err
	}

	r.ingress = ingress

	return r, nil
}

func (r *Room) AddParticipant(egress *PeerConnectionEgress) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.participants[egress.ID] = egress

	//go func() {
	buf := r.kfBuffer.GetPackets()
	for _, packet := range buf {
		egress.WriteRTP(packet)
		log.Printf("Buffer pkt %d", packet.SequenceNumber)
	}
	//}()
}

func (r *Room) RemoveParticipant(egress *PeerConnectionEgress) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.participants, egress.ID)
}

func (r *Room) Receive(b []byte) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.participants {
		err := p.Write(b)
		if err != nil {
			log.Printf("Unable to write data to PeerConnectionEgress. Error: %s", err)
		}
	}

	return nil
}

func (r *Room) ReceiveRTP(pkt *rtp.Packet) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	oldTs := pkt.Timestamp
	if r.lastTimeStamp == 0 {
		pkt.Timestamp = 0
	} else {
		pkt.Timestamp -= r.lastTimeStamp
	}
	r.lastTimeStamp = oldTs

	r.kfBuffer.Put(pkt)

	for _, p := range r.participants {
		err := p.WriteRTP(pkt)
		if err != nil {
			log.Printf("Unable to write data to PeerConnectionEgress. Error: %s", err)
		}
	}

	return nil
}

func (r *Room) Port() int {
	return r.ingress.Port()
}

func (r *Room) Has(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, has := r.participants[id]
	return has
}

func (r *Room) Participants() []*PeerConnectionEgress {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]*PeerConnectionEgress, 0, len(r.participants))

	for _, p := range r.participants {
		res = append(res, p)
	}

	return res
}
