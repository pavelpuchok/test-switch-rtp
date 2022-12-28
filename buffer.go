package main

import (
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"log"
	"sync"
)

type Buffer struct {
	vp8                   *codecs.VP8Packet
	buf                   []*rtp.Packet
	lastPkt               *rtp.Packet
	receivedPartitionHead bool
	mu                    sync.RWMutex
}

func NewBuffer(alloc int) *Buffer {
	return &Buffer{
		vp8: &codecs.VP8Packet{},
		buf: make([]*rtp.Packet, 0, alloc),
	}
}

func (b *Buffer) Put(p *rtp.Packet) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, err := b.vp8.Unmarshal(p.Payload)
	if err != nil {
		log.Printf("Unable to parse VP8 payload. Err: %s", err)
		return false
	}

	// Collects keyframes in buffer
	// https://www.rfc-editor.org/rfc/rfc7741.html#section-4.5.1

	if b.vp8.S == 1 {
		if b.receivedPartitionHead == false {
			b.buf = make([]*rtp.Packet, 0, len(b.buf))
		}
		b.buf = append(b.buf, p)
		b.receivedPartitionHead = true
		return true
	}

	if p.Marker {
		b.buf = append(b.buf, p)
		b.receivedPartitionHead = false
		return true
	}

	if b.receivedPartitionHead {
		b.buf = append(b.buf, p)
		return true
	}

	return false
}

func (b *Buffer) GetPackets() []*rtp.Packet {
	b.mu.RLock()
	defer b.mu.RUnlock()

	res := make([]*rtp.Packet, len(b.buf))
	for i, p := range b.buf {
		res[i] = p
	}
	return res
}
