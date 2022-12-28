package main

import (
	"github.com/google/uuid"
	"github.com/pion/interceptor"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
	"log"
	"sync"
)

type PeerConnectionEgress struct {
	ID             string
	peerConnection *webrtc.PeerConnection
	vTrack         *webrtc.TrackLocalStaticRTP
	rtpSender      *webrtc.RTPSender

	mu        sync.RWMutex
	sequence  uint16
	timestamp uint32

	started bool

	lastSSRC uint32
	lastTS   uint32
	lastSeq  uint16
}

var api *webrtc.API

func NewPeerConnectionEgress(offer webrtc.SessionDescription) (*PeerConnectionEgress, error) {
	ID := uuid.NewString()

	if api == nil {
		engine := &webrtc.MediaEngine{}
		err := engine.RegisterDefaultCodecs()
		if err != nil {
			return nil, err
		}

		i := &interceptor.Registry{}
		if err := webrtc.ConfigureRTCPReports(i); err != nil {
			return nil, err
		}

		//if err := webrtc.ConfigureTWCCSender(engine, i); err != nil {
		//	return nil, err
		//}

		api = webrtc.NewAPI(webrtc.WithMediaEngine(engine))
	}

	peerConnection, err := api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				//URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8}, "video", "pion")
	if err != nil {
		return nil, err
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		return nil, err
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
				return
			}
		}
	}()

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		log.Printf("PC: %s. Connection State has changed %s \n", ID, connectionState.String())

		if connectionState == webrtc.ICEConnectionStateFailed {
			if closeErr := peerConnection.Close(); closeErr != nil {
				log.Printf("PC: %s. Cannot close PC after ICE connection failed. %s", ID, closeErr)
			}
		}
	})

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		log.Printf("PC: %s. Incorrect SDP Offer received %s", ID, err)
		return nil, err
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Create Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}

	<-gatherComplete

	egress := &PeerConnectionEgress{
		ID:             ID,
		peerConnection: peerConnection,
		vTrack:         videoTrack,
		rtpSender:      rtpSender,
	}

	return egress, nil
}

func (p *PeerConnectionEgress) Write(b []byte) error {
	_, err := p.vTrack.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func (p *PeerConnectionEgress) WriteRTP(pkt *rtp.Packet) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.timestamp += pkt.Timestamp
	pkt.Timestamp = p.timestamp
	pkt.SequenceNumber = p.sequence
	p.sequence++

	return p.vTrack.WriteRTP(pkt)
}

func (p PeerConnectionEgress) LocalDescription() *webrtc.SessionDescription {
	return p.peerConnection.LocalDescription()
}
