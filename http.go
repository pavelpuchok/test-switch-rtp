package main

import (
	"encoding/json"
	"github.com/pion/webrtc/v3"
	"io"
	"log"
	"net/http"
)

type CreateRoomPayload struct {
}

type CreateRoomResponsePayload struct {
	ID   string `json:"id"`
	Port int    `json:"port"`
}

type CreatePeerConnectionPayload struct {
	Offer webrtc.SessionDescription `json:"offer"`
}

type CreatePeerConnectionResponsePayload struct {
	ID     string                     `json:"id"`
	Answer *webrtc.SessionDescription `json:"answer"`
}

type ListRoomsResponsePayload struct {
	Rooms []ListRoomsResponseRoomPayload `json:"rooms"`
}

type ListRoomsResponseRoomPayload struct {
	ID           string                                    `json:"id"`
	Port         int                                       `json:"port"`
	Participants []ListRoomsResponseRoomParticipantPayload `json:"participants"`
}

type ListRoomsResponseRoomParticipantPayload struct {
	ID string `json:"id"`
}

type SwitchPayload struct {
	PeerConnectionID string `json:"peerConnectionId"`
	RoomID           string `json:"roomId"`
}

type SwitchResponsePayload struct {
}

func NewHTTPServer(app *App) {
	mux := http.NewServeMux()

	mux.HandleFunc("/room", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.NotFound(w, req)
			return
		}

		payload := CreateRoomPayload{}
		raw, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = json.Unmarshal(raw, &payload)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		room, err := app.AddRoom(0)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		res := CreateRoomResponsePayload{
			ID:   room.ID,
			Port: room.Port(),
		}

		sendJSON(w, res)
	})

	mux.HandleFunc("/pc", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.NotFound(w, req)
			return
		}

		payload := CreatePeerConnectionPayload{}
		raw, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = json.Unmarshal(raw, &payload)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		egress, err := app.AddViewer(payload.Offer)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

		res := CreatePeerConnectionResponsePayload{
			ID:     egress.ID,
			Answer: egress.LocalDescription(),
		}

		sendJSON(w, res)
	})

	mux.HandleFunc("/rooms", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet {
			http.NotFound(w, req)
			return
		}

		rooms := app.ListRooms()
		res := ListRoomsResponsePayload{Rooms: make([]ListRoomsResponseRoomPayload, len(rooms))}

		for i, room := range rooms {
			ps := room.Participants()
			res.Rooms[i] = ListRoomsResponseRoomPayload{
				ID:           room.ID,
				Port:         room.Port(),
				Participants: make([]ListRoomsResponseRoomParticipantPayload, len(ps)),
			}

			for j, p := range ps {
				res.Rooms[i].Participants[j] = ListRoomsResponseRoomParticipantPayload{ID: p.ID}
			}
		}

		sendJSON(w, res)
	})

	mux.HandleFunc("/switch", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.NotFound(w, req)
			return
		}

		payload := SwitchPayload{}
		raw, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		err = json.Unmarshal(raw, &payload)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		app.Switch(payload.PeerConnectionID, payload.RoomID)

		res := SwitchResponsePayload{}

		sendJSON(w, res)
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	log.Println("Starting a HTTP Server. Visit http://localhost:3000")
	err := http.ListenAndServe(":3000", mux)
	if err != nil {
		panic(err)
		return
	}
}

func sendJSON(w http.ResponseWriter, v any) {
	resRaw, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(resRaw)
}
