package xatum

const (
	PacketC2S_Handshake = "shake"
	PacketS2C_Job       = "job"
	PacketC2S_Submit    = "submit"
	PacketS2C_Success   = "success"
	PacketS2C_Print     = "print"
	PacketS2C_Ping      = "ping"
	PacketC2S_Pong      = "pong"
)

type C2S_Handshake struct {
	Addr  string   `json:"addr"`  // wallet address
	Work  string   `json:"work"`  // worker name, by default "x"
	Agent string   `json:"agent"` // the mining software
	Algos []string `json:"algos"` // list of supported algorithms
}

type S2C_Job struct {
	Diff uint64 `json:"diff"` // difficulty of the job
	Blob B64    `json:"blob"` // xelis blob, which embeds work hash, extra nonce and public key (96 bytes) encoded as base64 string
}

type C2S_Submit struct {
	Data B64    `json:"data"` // the 112-bytes BlockMiner encoded as hex string
	Hash string `json:"hash"` // the 32-bytes PoW hash of BlockMiner encoded as hex string
}

type S2C_Success struct {
	Msg string `json:"msg"` // "ok" if share is good, otherwise msg contains the error message
}

type S2C_Print struct {
	Msg string `json:"msg"`
	Lvl uint8  `json:"lvl"` // log level, 0: verbose, 1: info, 2: warn, 3: error
}
