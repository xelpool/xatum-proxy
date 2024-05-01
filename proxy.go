package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
	"xatum-proxy/config"
	"xatum-proxy/log"
	"xatum-proxy/xatum"
	"xatum-proxy/xatum/client"
	"xatum-proxy/xelisutil"
)

const VERSION = "0.1.4"

// Job is a fast & efficient struct used for storing a job in memory
type Job struct {
	Blob   xelisutil.BlockMiner
	Diff   uint64
	Target [32]byte
}

var cl *client.Client

var sharesToPool = make(chan xatum.C2S_Submit, 1)

func main() {
	walletAddr := ""
	debug := false
	flag.StringVar(&walletAddr, "wallet", "", "your xelis address")
	flag.BoolVar(&debug, "debug", false, "true if you want to make logs verbose")
	flag.Parse()

	if debug {
		Cfg.Debug = true
	}

	if walletAddr != "" {
		Cfg.WalletAddress = walletAddr
	}

	if Cfg.WalletAddress == "YOUR WALLET ADDRESS HERE" {
		Cfg.WalletAddress = StringPrompt("Enter your wallet address:")

		if len(Cfg.WalletAddress) > 10 {
			saveCfg()
		} else {
			log.Err("invalid wallet address")
			os.Exit(0)
		}
	}

	log.Title("")
	log.Title(log.Bold + "          XATUM-PROXY v" + VERSION)
	log.Title(log.Purple + " (c) 2024 XelPool, licensed under MIT")
	log.Title("")
	log.Title(log.Reset+log.Cyan+" OS:", runtime.GOOS, "- arch:", runtime.GOARCH, "- threads:", runtime.NumCPU())
	log.Title("")

	go listenXatum()
	go listenGetwork()

	clientHandler()
}

func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

func clientHandler() {
	for {
		log.Debug("Starting new connection to the pool")

		var err error
		cl, err = client.NewClient(Cfg.PoolAddress)
		if err != nil {
			log.Errf("%v", err)
			time.Sleep(time.Second)
			continue
		}

		cl.Lock()
		err = cl.Send(xatum.PacketC2S_Handshake, xatum.C2S_Handshake{
			Addr:  Cfg.WalletAddress,
			Work:  "x",
			Agent: "XelMiner ALPHA",
			Algos: []string{config.ALGO},
		})
		cl.Unlock()
		if err != nil {
			log.Err(err)
			time.Sleep(time.Second)
			continue
		}

		log.Debug("sent handshake")

		go recvShares(cl)
		go readjobs(cl.Jobs)

		cl.Connect()

		time.Sleep(time.Second)
	}
}

var curJob Job
var mutCurJob sync.RWMutex

func recvShares(cl *client.Client) {
	log.Debug("recvShares started")
	for {
		share, ok := <-sharesToPool
		if !ok {
			log.Warn("sharesToPool chan closed")
			return
		}

		log.Info("share found, submitting to the pool")

		if !cl.Alive {
			log.Err("client is not alive")
			return
		}

		err := cl.Submit(share)
		if err != nil {
			log.Err("failed to submit share to pool:", err)
			return
		}
	}
}
func readjobs(clJobs chan xatum.S2C_Job) {
	for {
		job, ok := <-clJobs
		if !ok {
			return
		}

		mutCurJob.Lock()
		curJob = Job{
			Blob:   xelisutil.BlockMiner(job.Blob),
			Diff:   job.Diff,
			Target: xelisutil.GetTargetBytes(job.Diff),
		}
		mutCurJob.Unlock()

		log.Infof("new job with difficulty %d", job.Diff)
		log.Debugf("new job: diff %d, blob %x", job.Diff, job.Blob)

		go func() {
			srv.RLock()
			defer srv.RUnlock()

			for _, v := range srv.Connections {
				SendJob(v, job.Diff, job.Blob)
			}
		}()

		go sendJobToWebsocket(job.Diff, job.Blob)

	}
}
