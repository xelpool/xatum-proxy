package server

import (
	"crypto/tls"
	"encoding/json"
	"net"
	"strconv"
	"sync"
	"time"
	"xatum-proxy/log"
	"xatum-proxy/util"
	"xatum-proxy/xatum"
	"xatum-proxy/xelisutil"
)

const MAX_CONNECTIONS_PER_IP = 100

type Server struct {
	Connections []*Connection
	connsPerIp  map[string]uint32

	NewConnections chan *Connection

	sync.RWMutex
}

type Connection struct {
	Conn net.Conn
	Id   uint64

	CurrentJob ConnJob
	LastJob    ConnJob

	LastShare time.Time // in unix milliseconds
	Score     int32
	Wallet    string

	sync.RWMutex
}

type ConnJob struct {
	Diff uint64

	BlockMiner xelisutil.BlockMiner

	SubmittedNonces []uint64
}

func (c *Connection) Send(name string, a any) error {
	data, err := json.Marshal(a)
	if err != nil {
		panic(err)
	}
	return c.SendBytes(append([]byte(name+"~"), data...))
}
func (c *Connection) SendBytes(data []byte) error {
	log.Net(">>>", string(data))
	c.Conn.SetWriteDeadline(time.Now().Add(20 * time.Second))
	_, err := c.Conn.Write(append(data, '\n'))
	if err != nil {
		return err
	}
	return nil
}
func (c *Connection) SendJob(job xatum.S2C_Job) {
	c.Send(xatum.PacketS2C_Job, job)
}

func (s *Server) Start(port uint16) {
	s.NewConnections = make(chan *Connection, 1)
	s.connsPerIp = make(map[string]uint32, 100)

	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Info("generating a new TLS certificate: no cert file found:", err)

		certPem, keyPem, err := GenCertificate()
		if err != nil {
			log.Fatal(err)
		}

		cert, err = tls.X509KeyPair(certPem, keyPem)
		if err != nil {
			log.Fatal(err)
		}

	}

	listener, err := tls.Listen("tcp", "0.0.0.0:"+strconv.FormatUint(uint64(port), 10), &tls.Config{
		Certificates: []tls.Certificate{
			cert,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Xatum server listening on port", port)

	for {
		c, err := listener.Accept()
		if err != nil {
			log.Err(err)
			continue
		}
		minerIp := util.RemovePort(c.RemoteAddr().String())

		log.Debug("new incoming connection with IP", minerIp)

		conn := &Connection{
			Conn:      c,
			Id:        util.RandomUint64(),
			LastShare: time.Now(),
		}
		go s.handleConnection(conn)
	}
}

// Server MUST be locked before calling this
func (s *Server) Kick(id uint64) {
	var connectionsNew = make([]*Connection, 0, len(s.Connections))

	for _, v := range s.Connections {
		if v.Id == id {
			v.Conn.Close()

			ipAddr := util.RemovePort(v.Conn.RemoteAddr().String())

			if s.connsPerIp[ipAddr] > 0 {
				s.connsPerIp[ipAddr]--
			}

		} else {
			connectionsNew = append(connectionsNew, v)
		}
	}
	s.Connections = connectionsNew
}

// this function locks Server
func (srv *Server) handleConnection(conn *Connection) {
	log.Dev("handling connection with ID", conn.Id)

	srv.Lock()
	defer srv.Unlock()

	ipAddr := util.RemovePort(conn.Conn.RemoteAddr().String())

	if srv.connsPerIp[ipAddr] > MAX_CONNECTIONS_PER_IP {
		log.Debug("address", ipAddr, "reached connections per IP limit")
		conn.Conn.Close()
		return
	}

	srv.connsPerIp[ipAddr]++

	srv.Connections = append(srv.Connections, conn)
	log.Debug("handling connection")

	srv.NewConnections <- conn
}
