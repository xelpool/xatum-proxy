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

type GetworkConn struct {
	conn *websocket.Conn

	sync.RWMutex
}

// GetworkConn MUST be locked before calling this
func (g *GetworkConn) WriteJSON(data interface{}) error {
	return g.conn.WriteJSON(data)
}

func (g *GetworkConn) IP() string {
	return g.conn.RemoteAddr().String()
}

func (g *GetworkConn) Close() error {
	return g.conn.Close()
}

var socketsMut sync.RWMutex
var sockets []*GetworkConn

// sends a job to all the websockets, and removes old websockets
func sendJobToWebsocket(diff uint64, blob []byte) {
	log.Dev("sendJobToWebsocket: num sockets:", len(sockets))

	socketsMut.Lock()
	defer socketsMut.Unlock()

	log.Dev("sendJobToWebsocket: socketsMut Lock success")

	// remove disconnected sockets

	sockets2 := make([]*GetworkConn, 0, len(sockets))
	for _, c := range sockets {
		if c == nil {
			continue
		}
		sockets2 = append(sockets2, c)
	}
	log.Dev("sendJobToWebsocket: going from", len(sockets), "to", len(sockets2), "getwork miners")
	sockets = sockets2

	if len(sockets) > 0 {
		log.Info("Sending job to", len(sockets), "GetWork miners")
	}

	// send jobs to the remaining sockets

	for ix, cx := range sockets {
		if cx == nil {
			log.Dev("cx is nil")
			continue
		}

		i := ix
		c := cx

		// send job in a new thread to avoid blocking the main thread and reduce latency
		go func() {
			log.Debug("sendJobToWebsocket: sending to IP", c.IP())

			c.Lock()
			err := c.WriteJSON(map[string]any{
				"new_job": BlockTemplate{
					Difficulty: strconv.FormatUint(diff, 10),
					TopoHeight: 0,
					Template:   hex.EncodeToString(blob),
				},
			})
			c.Unlock()

			log.Debug("sendJobToWebsocket: sent to IP", c.IP())

			// if write failed, close the connection (if it isn't already closed) and remove it from
			// the list of sockets
			if err != nil {
				log.Warn("sendJobToWebsocket: cannot send job:", err)
				c.Lock()
				c.Close()
				c.Unlock()

				socketsMut.Lock()
				sockets[i] = nil
				socketsMut.Unlock()

				log.Warn("sendJobToWebsocket: cannot send job DONE")
				return
			}

			c.RLock()
			log.Debug("sendJobToWebsocket: done, sent to IP", c.IP())
			c.RUnlock()
		}()
	}
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
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}
	defer conn.Close()

	log.Info("Miner with IP", conn.RemoteAddr().String(), "connected to Getwork")

	socketsMut.Lock()
	c := &GetworkConn{conn: conn}
	sockets = append(sockets, c)
	socketsMut.Unlock()

	// send first job
	mutCurJob.Lock()
	if curJob.Diff == 0 {
		log.Debug("not sending first job, because there is no first job yet")
		mutCurJob.Unlock()
		return
	}

	log.Debug("sending first job")

	diff := strconv.FormatUint(curJob.Diff, 10)
	blob := curJob.Blob
	mutCurJob.Unlock()

	c.Lock()
	err = c.WriteJSON(map[string]any{
		"new_job": BlockTemplate{
			Difficulty: diff,
			TopoHeight: 0,
			Template:   hex.EncodeToString(blob[:]),
		},
	})
	c.Unlock()
	if err != nil {
		log.Warn("failed to send first job:", err)
	}
	// done sending first job

	log.Debug("done sending first job")

	for {
		c.RLock()
		mt, message, err := c.conn.ReadMessage()
		c.RUnlock()
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
			if msgJson["block_template"] == nil {
				log.Debug("miner_work and block_template are nil")
				continue
			} else {
				msgJson["miner_work"] = msgJson["block_template"]
			}
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

		// send dummy "accepted" reply
		c.Lock()
		err = c.conn.WriteMessage(websocket.TextMessage, []byte(`"block_accepted"`))
		c.Unlock()
		if err != nil {
			log.Err("failed to send dummy accept reply:", err)
		}

		// send share to pool
		sharesToPool <- xatum.C2S_Submit{
			Data: minerBlob,
			Hash: hex.EncodeToString(pow[:]),
		}
	}
}
