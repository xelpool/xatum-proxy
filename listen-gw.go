package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"net/http"
	"strconv"
	"sync"
	"xatum-proxy/log"
	"xatum-proxy/xatum"
	"xatum-proxy/xelishash"
	"xatum-proxy/xelisutil"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func fmtMessageType(mt int) string {
	if mt == websocket.BinaryMessage {
		return "binary"
	} else if mt == websocket.TextMessage {
		return "text"
	} else {
		return "Unknown Message Type"
	}
}

var socketsMut sync.RWMutex
var sockets []*websocket.Conn

// sends a job to all the websockets, and removes old websockets
func sendJobToWebsocket(diff uint64, blob []byte) {
	socketsMut.Lock()
	defer socketsMut.Unlock()

	sockets2 := make([]*websocket.Conn, 0, len(sockets))

	for _, c := range sockets {
		if c == nil {
			continue
		}

		err := c.WriteJSON(map[string]any{
			"new_job": BlockTemplate{
				Difficulty: strconv.FormatUint(diff, 10),
				TopoHeight: 0,
				Template:   hex.EncodeToString(blob),
			},
		})
		if err != nil {
			log.Warn(err)
			continue
		}

		sockets2 = append(sockets2, c)
	}
	sockets = sockets2

}

func listenGetwork() {
	flag.Parse()

	http.HandleFunc("/", wsHandler)

	ip := "0.0.0.0:" + strconv.FormatUint(uint64(Cfg.GetworkBindPort), 10)

	log.Info("Getwork server listening on port", Cfg.GetworkBindPort)

	log.Fatal(http.ListenAndServe(ip, nil))
}

type BlockTemplate struct {
	Difficulty string `json:"difficulty"`
	Height     uint64 `json:"height"`
	TopoHeight uint64 `json:"topoheight"`
	Template   string `json:"template"`
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}
	defer c.Close()

	socketsMut.Lock()
	sockets = append(sockets, c)
	socketsMut.Unlock()

	// send first job
	mutCurJob.Lock()
	if curJob.Diff == 0 {
		log.Debug("not sending first job, because there is no first job yet")
		mutCurJob.Unlock()
		return
	}
	diff := strconv.FormatUint(curJob.Diff, 10)
	blob := curJob.Blob
	mutCurJob.Unlock()
	c.WriteJSON(map[string]any{
		"new_job": BlockTemplate{
			Difficulty: diff,
			TopoHeight: 0,
			Template:   hex.EncodeToString(blob[:]),
		},
	})
	// done sending first job

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Info("Getwork miner disconnected:", err)
			break
		}

		log.Debugf("recv: %s, type: %s", message, fmtMessageType(mt))

		var msgJson map[string]any

		err = json.Unmarshal([]byte(message), &msgJson)

		if err != nil {
			log.Err(err)
		}

		if msgJson["miner_work"] == nil {
			log.Debug("miner_work is nil")
			continue
		}

		minerWork := msgJson["miner_work"].(string)

		minerBlob, err := hex.DecodeString(minerWork)
		if err != nil {
			log.Err(err)
			continue
		}

		if len(minerBlob) != xelisutil.BLOCKMINER_LENGTH {
			log.Info()
			continue
		}

		blob := xelisutil.BlockMiner(minerBlob)

		// calculate PoW (unfortunatly it's needed)
		scratchpad := xelishash.ScratchPad{}
		pow := blob.PowHash(&scratchpad)

		// send share to pool
		sharesToPool <- xatum.C2S_Submit{
			Data: minerBlob,
			Hash: hex.EncodeToString(pow[:]),
		}

		/*err = c.WriteMessage(mt, message)
		if err != nil {
			log.Info("write:", err)
			break
		}*/
	}
}
