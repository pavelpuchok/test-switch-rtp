package main

import (
	"github.com/pion/rtp"
	"log"
	"net"
)

type Receiver interface {
	Receive(raw []byte) error
	ReceiveRTP(pkt *rtp.Packet) error
}

type UDPIngress struct {
	s    *net.UDPConn
	r    Receiver
	addr *net.UDPAddr
}

func NewUDPIngress(w Receiver, port int) (*UDPIngress, error) {
	addr := &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: port}

	s, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	u := &UDPIngress{
		s:    s,
		r:    w,
		addr: s.LocalAddr().(*net.UDPAddr),
	}

	u.listen()

	return u, nil
}

func (s *UDPIngress) listen() {
	go func() {
		var b = make([]byte, 1500)
		pkt := rtp.Packet{}
		for {
			n, _, err := s.s.ReadFromUDP(b)
			if err != nil {
				return
			}

			err = pkt.Unmarshal(b[:n])
			if err != nil {
				log.Printf("Unable to unmarshal packet. Error: %s", err)
				return
			}

			err = s.r.ReceiveRTP(&pkt)
			if err != nil {
				log.Printf("Unable to receive packet. Receiver error: %s", err)
				return
			}
		}
	}()
}

func (s *UDPIngress) Port() int {
	return s.addr.Port
}
