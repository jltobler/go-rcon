package packet

import "testing"

func TestMarshalAndUnmarshal(t *testing.T) {
	p := New(1, Command, "test")

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
