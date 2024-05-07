package client

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"net"
	"strings"
	"sync"
	"time"
	"xatum-proxy/log"
	"xatum-proxy/xatum"
)

type Client struct {
	PoolAddress string
	conn        net.Conn

	Alive bool

	LastJob time.Time

	Jobs    chan xatum.S2C_Job
	Prints  chan xatum.S2C_Print
	Success chan xatum.S2C_Success

	sync.RWMutex
}

func NewClient(poolAddr string) (*Client, error) {
	cl := &Client{
		PoolAddress: poolAddr,

		Jobs:    make(chan xatum.S2C_Job, 1),
		Prints:  make(chan xatum.S2C_Print, 1),
		Success: make(chan xatum.S2C_Success, 1),
	}

	conn, err := tls.Dial("tcp", cl.PoolAddress, &tls.Config{
		InsecureSkipVerify: true,
	})
	cl.conn = conn
	if err != nil {
		log.Warnf("connection failed: %s", err)
		return nil, err
	}

	cl.Alive = true

	return cl, nil
}

// Client must NOT be locked before calling this
func (cl *Client) Connect() {
	rdr := bufio.NewReader(cl.conn)
	for {
		cl.conn.SetReadDeadline(time.Now().Add(time.Minute))

		str, err := rdr.ReadString('\n')

		if err != nil {
			log.Warnf("connection closed: %s", err)
			cl.Close()
			return
		}
		log.Net("<<<", str)

		spl := strings.SplitN(str, "~", 2)
		if spl == nil || len(spl) < 2 {
			log.Warn("packet data is malformed")
			continue
		}

		pack := spl[0]

		cl.Lock()
		if pack != xatum.PacketS2C_Job && time.Since(cl.LastJob) > time.Minute {
			log.Err("no jobs received in the last minute, reconnecting")

			cl.Close()
			cl.Unlock()
			return
		}
		cl.Unlock()

		if pack == xatum.PacketS2C_Job {
			pData := xatum.S2C_Job{}

			err := json.Unmarshal([]byte(spl[1]), &pData)
			if err != nil {
				log.Warn("failed to parse data")
				cl.Close()
				return
			}

			log.Debug("ok, job received, sending to channel")

			cl.Jobs <- pData
			cl.Lock()
			cl.LastJob = time.Now()
			cl.Unlock()

			log.Debug("ok, done sending to channel")
		} else if pack == xatum.PacketS2C_Print {
			pData := xatum.S2C_Print{}
			err := json.Unmarshal([]byte(spl[1]), &pData)
			if err != nil {
				log.Warn("failed to parse data")
				cl.Close()
				return
			}

			const PREFIX = "message from pool:"

			switch pData.Lvl {
			case 1:
				log.Infof(PREFIX+" %s", pData.Msg)
			case 2:
				log.Warnf(PREFIX+" %s", pData.Msg)
			case 3:
				log.Errf(PREFIX+" %s", pData.Msg)
			}

		} else if pack == xatum.PacketS2C_Ping {
			cl.Send("pong", map[string]any{})
		} else {
			log.Warnf("Unknown packet %s", pack)
		}

	}

}

// Client MUST be locked before calling this
func (cl *Client) Close() {
	err := cl.conn.Close()
	if err != nil {
		log.Debug("cl.conn.Close failed:", err)
	}
	close(cl.Jobs)
	close(cl.Prints)
	close(cl.Success)
	cl.Alive = false
}

// Client MUST be locked before calling this
func (c *Client) Send(name string, a any) error {
	data, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	c.SendBytes(append([]byte(name+"~"), data...))
	return nil
}

// Client MUST be locked before calling this
func (cl *Client) SendBytes(data []byte) error {
	log.Net(">>>", string(data))

	_, err := cl.conn.Write(append(data, '\n'))
	if err != nil {
		return err
	}

	return nil
}

func (cl *Client) Submit(pack xatum.C2S_Submit) error {

	return cl.Send(xatum.PacketC2S_Submit, pack)
}
