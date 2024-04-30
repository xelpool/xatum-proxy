package xatum

import (
	"encoding/json"
	"errors"
	"strings"
)

type Packet struct {
	Name string
	Data any
}

func NewPacket(name string, data any) Packet {
	return Packet{
		Name: name,
		Data: data,
	}
}
func NewPacketFromString(data string, pack *Packet) error {
	spl := strings.SplitN(data, "~", 2)
	if spl == nil || len(spl) < 2 {
		return errors.New("malformed packet string")
	}

	pack.Name = spl[0]

	err := json.Unmarshal([]byte(spl[1]), &pack.Data)

	return err
}

func (p Packet) ToString() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return p.Name + "~" + string(data), nil
}
