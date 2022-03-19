package packet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"
	"unicode"
)

// Enumerates Packet types
const (
	Response Kind = iota
	_
	Command
	Login
	_
	Termination
)

type Kind uint32

// Packet defines RCON protocol encoding
type Packet struct {
	Length  uint32
	ID      uint32
	Kind    Kind
	Payload string
}

// New creates and returns a Packet
func New(id uint32, kind Kind, payload string) *Packet {
	return &Packet{
		Length:  uint32(len(payload)+10),
		ID:      id,
		Kind:    kind,
		Payload: payload,
	}
}

// Marshal returns the RCON encoding of Packet p. If Packet p
// is nil or contains invalid payload, Marshal returns an error.
func Marshal(p *Packet) ([]byte, error) {
	if p == nil {
		return nil, errors.New("nil packet provided")
	}

	for i := 0; i < len(p.Payload); i++ {
		if p.Payload[i] > unicode.MaxASCII {
			return nil, errors.New("payload contains non-ASCII characters")
		}
	}

	buf := bytes.Buffer{}
	b := make([]byte, 4)

	binary.LittleEndian.PutUint32(b, p.Length)
	buf.Write(b)

	binary.LittleEndian.PutUint32(b, p.ID)
	buf.Write(b)

	binary.LittleEndian.PutUint32(b, uint32(p.Kind))
	buf.Write(b)

	buf.WriteString(p.Payload)

	buf.Write([]byte{0,0})

	return buf.Bytes(), nil
}

// Unmarshal parses the RCON encoded Packet data and stores the
// result in the value pointed to by p. If p is nil or data
// contains invalid RCON encoding, Unmarshal returns an error.
func Unmarshal(data []byte, p *Packet) error {
	if p == nil {
		return errors.New("nil packet provided")
	}

	if len(data) < 14 || data[len(data)-1] != 0 || data[len(data)-2] != 0 {
		return errors.New("invalid packet bytes")
	}

	length := binary.LittleEndian.Uint32(data[0:4])
	if uint32(len(data)) != length+4 {
		return errors.New("incorrect packet length")
	}

	id := binary.LittleEndian.Uint32(data[4:8])

	var kind Kind
	switch binary.LittleEndian.Uint32(data[8:12]) {
	case 0:
		kind = Response
	case 2:
		kind = Command
	case 3:
		kind = Login
	case 5:
		kind = Termination
	default:
		return errors.New("invalid packet type")
	}

	buf := strings.Builder{}
	b := data[12:len(data)-2]
	for i := 0; i < len(b); i++ {
		if b[i] > unicode.MaxASCII {
			return errors.New("payload contains non-ASCII characters")
		}
		buf.WriteByte(b[i])
	}
	payload := buf.String()

	p.Length = length
	p.ID = id
	p.Kind = kind
	p.Payload = payload

	return nil
}
