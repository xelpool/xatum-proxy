package main

import (
	"encoding/json"
	"os"
	"xatum-proxy/log"
)

type Config struct {
	WalletAddress   string
	PoolAddress     string
	XatumBindPort   uint16
	GetworkBindPort uint16
	Debug           bool
}

// 5210: Getwork
// 5211: Xatum
// 5212: Xatum public (mining pools)

var Cfg = Config{
	Debug:           false,
	WalletAddress:   "YOUR WALLET ADDRESS HERE",
	PoolAddress:     "auto.xatum.xelpool.com:5212",
	XatumBindPort:   5211,
	GetworkBindPort: 5210,
}

func init() {
	loadCfg()

	if Cfg.Debug {
		log.LogLevel = 2
	}
}

func loadCfg() {
	data, err := os.ReadFile(path() + "/config.json")

	if err != nil {
		log.Warn("failed to open configuration:", err)
		saveCfg()
		return
	}

	err = json.Unmarshal(data, &Cfg)

	if err != nil {
		log.Warn("failed to decode configuration:", err)
		return
	}
}

func saveCfg() {
	data, err := json.MarshalIndent(Cfg, "", "\t")
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(path()+"/config.json", data, 0o666)
	if err != nil {
		log.Err(err)
	}
}

func path() string {
	return "."
	/*ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	return strings.TrimSuffix(exPath, string(os.PathSeparator))*/
}
