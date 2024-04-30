package main

import (
	"bufio"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
	"xatum-proxy/config"
	"xatum-proxy/log"
	"xatum-proxy/xatum"
	"xatum-proxy/xatum/server"
	"xatum-proxy/xelisutil"
)

var srv = &server.Server{}

func listenXatum() {
	go waitConnections(srv)

	srv.Start(Cfg.XatumBindPort)

}

func waitConnections(srv *server.Server) {
	for {
		conn := <-srv.NewConnections

		go handleConn(srv, conn)

	}
}

func handleConn(s *server.Server, conn *server.Connection) {
	log.Dev("handleConn")
	rdr := bufio.NewReader(conn.Conn)

	packetsRecv := 0

	go sendPingPackets(s, conn)

	for {
		conn.Lock()
		conn.Conn.SetReadDeadline(time.Now().Add(config.SLAVE_MINER_TIMEOUT * time.Second))
		conn.Unlock()

		str, err := rdr.ReadString('\n')

		packetsRecv++

		if err != nil {
			s.Lock() // TODO: put this Lock in pool too
			defer s.Unlock()

			s.Kick(conn.Id)
			return
		}

		log.Net("<<<", str)

		err = handleConnPacket(s, conn, str, packetsRecv)
		if err != nil {
			log.Err(err)
			return
		}
	}
}

func sendPingPackets(s *server.Server, conn *server.Connection) {
	for {
		time.Sleep((config.SLAVE_MINER_TIMEOUT - 5) * time.Second)

		err := conn.Send(xatum.PacketS2C_Ping, map[string]any{})
		if err != nil {
			log.Warn(err)
			conn.Conn.Close()
			s.Kick(conn.Id)
			return
		}
	}
}

func handleConnPacket(s *server.Server, conn *server.Connection, str string, packetsRecv int) error {

	conn.Lock()
	defer conn.Unlock()

	spl := strings.SplitN(str, "~", 2)
	if spl == nil || len(spl) < 2 {
		log.Warn("packet data is malformed, spl:", spl)
		conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
			Msg: "malformed packet data",
			Lvl: 3,
		})
		return nil
	}

	pack := spl[0]

	if packetsRecv == 1 && spl[0] != xatum.PacketC2S_Handshake {
		err := fmt.Errorf("first packet must be %s, got %s", xatum.PacketC2S_Handshake, spl[0])
		conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
			Msg: "first packet must be a handshake",
			Lvl: 3,
		})
		s.Kick(conn.Id)
		return err
	}

	if pack == xatum.PacketC2S_Handshake {
		if packetsRecv != 1 {
			err := fmt.Errorf("client sent more than one handshake")
			conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
				Msg: "more than one handshake received",
				Lvl: 3,
			})
			s.Kick(conn.Id)
			return err
		}

		pData := xatum.C2S_Handshake{}

		err := json.Unmarshal([]byte(spl[1]), &pData)
		if err != nil {
			err = fmt.Errorf("failed to parse data: %v", err)
			conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
				Msg: "failed to parse data",
				Lvl: 3,
			})
			s.Kick(conn.Id)
			return err
		}

		if !slices.Contains(pData.Algos, config.ALGO) {
			err := fmt.Errorf("miner does not support algorithm %s", config.ALGO)
			conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
				Msg: "your miner does not support algorithm " + config.ALGO,
				Lvl: 3,
			})
			s.Kick(conn.Id)
			return err
		}

		log.Infof("New miner | Address: %s %s UserAgent: %s Algos: %s", pData.Addr, pData.Work, pData.Agent, pData.Algos)

		conn.Wallet = pData.Addr

		// send first job

		mutCurJob.Lock()

		if curJob.Diff == 0 {
			log.Debug("not sending first job, because there is no first job yet")
			mutCurJob.Unlock()
			return nil
		}

		diff := curJob.Diff
		blob := curJob.Blob

		mutCurJob.Unlock()

		log.Debugf("first job diff %d blob %x", diff, blob)

		SendJob(conn, diff, blob[:])
	} else if pack == xatum.PacketC2S_Pong {
		log.Dev("received pong packet")
	} else if pack == xatum.PacketC2S_Submit {

		pData := xatum.C2S_Submit{}

		err := json.Unmarshal([]byte(spl[1]), &pData)
		if err != nil {
			err := fmt.Errorf("failed to parse data")
			conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
				Msg: "failed to parse data",
				Lvl: 3,
			})
			s.Kick(conn.Id)
			return err
		}

		// send the share to pool

		log.Dev("sending share to the pool")
		sharesToPool <- pData
	} else {
		err := fmt.Errorf("unknown packet %s", pack)
		conn.Send(xatum.PacketS2C_Print, xatum.S2C_Print{
			Msg: "unknown packet " + pack,
			Lvl: 3,
		})
		s.Kick(conn.Id)
		return err
	}

	return nil
}

// Sends job to a miner connected to the proxy
// NOTE: Connection MUST be locked before calling this
func SendJob(v *server.Connection, blockDiff uint64, blob []byte) {
	v.LastJob = v.CurrentJob

	blMiner := xelisutil.BlockMiner(blob)

	v.CurrentJob = server.ConnJob{
		Diff:            blockDiff,
		BlockMiner:      blMiner,
		SubmittedNonces: make([]uint64, 0, 8),
	}

	nonceExtra := blMiner.GetExtraNonce()

	rnd := make([]byte, 4)
	_, err := rand.Read(rnd)
	if err != nil {
		log.Fatal(err)
	}

	nonceExtra[28] = rnd[0]
	nonceExtra[29] = rnd[1]
	nonceExtra[30] = rnd[2]
	nonceExtra[31] = rnd[3]

	blMiner.SetExtraNonce(nonceExtra)

	log.Devf("sending job to miner with ID %d", v.Id)
	v.SendJob(xatum.S2C_Job{
		Diff: blockDiff,
		Blob: blMiner[:],
	})
}
