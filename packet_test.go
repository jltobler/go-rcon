package rcon

import "testing"

func TestMarshalAndUnmarshal(t *testing.T) {
	p := NewPacket(CommandPacket, "test")

	data, err := Marshal(p)
	if err != nil {
		t.Fail()
	}

	r := &Packet{}

	if err := Unmarshal(data, r); err != nil {
		t.Fail()
	}

	if p.Length != r.Length || p.ID != r.ID || p.Kind != r.Kind || p.Payload != r.Payload {
		t.Fail()
	}
}
