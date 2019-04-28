package handler

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/giongto35/cloud-game/config"
	"github.com/giongto35/cloud-game/cws"
	"github.com/giongto35/cloud-game/handlers/client"
	"github.com/gorilla/websocket"
)

const (
	width        = 256
	height       = 240
	scale        = 3
	title        = "NES"
	gameboyIndex = "./static/gameboy.html"
	debugIndex   = "./static/index_ws.html"
)

var indexFN = gameboyIndex

// Time allowed to write a message to the peer.
var readWait = 30 * time.Second
var writeWait = 30 * time.Second

// Flag to determine if the server is overlord or not
var IsOverlord = false
var upgrader = websocket.Upgrader{}

// ID to peerconnection
//var peerconnections = map[string]*webrtc.WebRTC{}
var serverID = ""
var oclient *cws.Client

// getWeb returns web frontend
func getWeb(w http.ResponseWriter, r *http.Request) {
	bs, err := ioutil.ReadFile(indexFN)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(bs)
}

// Handle normal traffic (from browser to host)
func ws(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("[!] WS upgrade:", err)
		return
	}
	defer c.Close()
	//var gameName string
	//var roomID string
	//var playerIndex int

	//// Create connection to overlord
	//client := NewClient(c)
	////sessionID := strconv.Itoa(rand.Int())
	//sessionID := uuid.Must(uuid.NewV4()).String()

	//wssession := &Session{
	//client:         client,
	//peerconnection: webrtc.NewWebRTC(),
	//// The server session is maintaining
	//}

	//client.send(WSPacket{
	//ID:   "gamelist",
	//Data: getEncodedGameList(),
	//}, nil)

	//client.receive("heartbeat", func(resp WSPacket) WSPacket {
	//return resp
	//})

	//client.receive("initwebrtc", func(resp WSPacket) WSPacket {
	//log.Println("Received user SDP")
	//localSession, err := wssession.peerconnection.StartClient(resp.Data, width, height)
	//if err != nil {
	//log.Fatalln(err)
	//}

	//return WSPacket{
	//ID:        "sdp",
	//Data:      localSession,
	//SessionID: sessionID,
	//}
	//})

	//client.receive("save", func(resp WSPacket) (req WSPacket) {
	//log.Println("Saving game state")
	//req.ID = "save"
	//req.Data = "ok"
	//if roomID != "" {
	//err = rooms[roomID].director.SaveGame()
	//if err != nil {
	//log.Println("[!] Cannot save game state: ", err)
	//req.Data = "error"
	//}
	//} else {
	//req.Data = "error"
	//}

	//return req
	//})

	//client.receive("load", func(resp WSPacket) (req WSPacket) {
	//log.Println("Loading game state")
	//req.ID = "load"
	//req.Data = "ok"
	//if roomID != "" {
	//err = rooms[roomID].director.LoadGame()
	//if err != nil {
	//log.Println("[!] Cannot load game state: ", err)
	//req.Data = "error"
	//}
	//} else {
	//req.Data = "error"
	//}

	//return req
	//})

	//client.receive("start", func(resp WSPacket) (req WSPacket) {
	//gameName = resp.Data
	//roomID = resp.RoomID
	//playerIndex = resp.PlayerIndex
	//isNewRoom := false

	//log.Println("Starting game")
	//// If we are connecting to overlord, request serverID from roomID
	//if oclient != nil {
	//roomServerID := getServerIDOfRoom(oclient, roomID)
	//log.Println("Server of RoomID ", roomID, " is ", roomServerID)
	//if roomServerID != "" && wssession.ServerID != roomServerID {
	//// TODO: Re -register
	//go bridgeConnection(wssession, roomServerID, gameName, roomID, playerIndex)
	//return
	//}
	//}

	//roomID, isNewRoom = startSession(wssession.peerconnection, gameName, roomID, playerIndex)
	//// Register room to overlord if we are connecting to overlord
	//if isNewRoom && oclient != nil {
	//oclient.send(WSPacket{
	//ID:   "registerRoom",
	//Data: roomID,
	//}, nil)
	//}
	//req.ID = "start"
	//req.RoomID = roomID
	//req.SessionID = sessionID

	//return req
	//})

	//client.receive("candidate", func(resp WSPacket) (req WSPacket) {
	//// Unuse code
	//hi := pionRTC.ICECandidateInit{}
	//err = json.Unmarshal([]byte(resp.Data), &hi)
	//if err != nil {
	//log.Println("[!] Cannot parse candidate: ", err)
	//} else {
	//// webRTC.AddCandidate(hi)
	//}
	//req.ID = "candidate"

	//return req
	//})

	client := client.NewBrowserClient(c)
	client.listen()
}

func getServerIDOfRoom(oc *Client, roomID string) string {
	log.Println("Request overlord roomID")
	packet := oc.syncSend(
		cws.WSPacket{
			ID:   "getRoom",
			Data: roomID,
		},
	)
	log.Println("Received roomID from overlord")

	return packet.Data
}

func bridgeConnection(session *Session, serverID string, gameName string, roomID string, playerIndex int) {
	log.Println("Bridging connection to other Host ", serverID)
	client := session.client
	// Ask client to init

	log.Println("Requesting offer to browser", serverID)
	resp := client.syncSend(cws.WSPacket{
		ID:   "requestOffer",
		Data: "",
	})

	log.Println("Sending offer to overlord to relay message to target host", resp.TargetHostID)
	// Ask overlord to relay SDP packet to serverID
	resp.TargetHostID = serverID
	remoteTargetSDP := oclient.syncSend(resp)
	log.Println("Got back remote host SDP, sending to browser")
	// Send back remote SDP of remote server to browser
	//client.syncSend(WSPacket{
	//ID:   "sdp",
	//Data: remoteTargetSDP.Data,
	//})
	client.send(cws.WSPacket{
		ID:   "sdp",
		Data: remoteTargetSDP.Data,
	}, nil)
	log.Println("Init session done, start game on target host")

	oclient.syncSend(cws.WSPacket{
		ID:           "start",
		Data:         gameName,
		TargetHostID: serverID,
		RoomID:       roomID,
		PlayerIndex:  playerIndex,
	})
	log.Println("Game is started on remote host")
}

func createOverlordConnection() (*websocket.Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(*config.OverlordHost, nil)
	if err != nil {
		return nil, err
	}

	return c, nil
}