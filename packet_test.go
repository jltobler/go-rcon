package rcon

import "testing"

func TestMarshalAndUnmarshal(t *testing.T) {
	p := NewPacket(CommandPacket, "test")

	data, err := Marshal(p, CharSetASCII)
	if err != nil {
		t.Fail()
	}

	r := &Packet{}

	if err := Unmarshal(data, r, CharSetASCII); err != nil {
		t.Fail()
	}

	if p.Length != r.Length || p.ID != r.ID || p.Kind != r.Kind || p.Payload != r.Payload {
		t.Fail()
	}
}

func TestMarshal_CharSets(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		charSet CharSet
		wantErr bool
	}{
		// --- CharSetASCII ---
		{
			name:    "ASCII - Valid Standard Text",
			payload: "Hello World 123",
			charSet: CharSetASCII,
			wantErr: false,
		},
		{
			name:    "ASCII - Invalid (Contains Latin-1 / Ä)",
			payload: "H\xc3\xa4llo", // 'ä' as UTF-8 byte sequence
			charSet: CharSetASCII,
			wantErr: true,
		},
		{
			name:    "ASCII - Invalid (Boundary byte 0x80)",
			payload: "test\x80", // 128 > MaxASCII (127)
			charSet: CharSetASCII,
			wantErr: true,
		},

		// --- CharSetLatin_1 (ISO-8859-1) ---
		{
			name:    "Latin1 - Valid Standard ASCII Range",
			payload: "Hello World",
			charSet: CharSetLatin1,
			wantErr: false,
		},
		{
			name:    "Latin1 - Valid Extended Range (ä, ö, ü as single bytes)",
			payload: "H\xe4llo", // 0xE4 = 'ä' in ISO-8859-1 (<= 255)
			charSet: CharSetLatin1,
			wantErr: false,
		},

		// --- CharSetUTF8 ---
		{
			name:    "UTF8 - Valid Plain ASCII",
			payload: "Hello World",
			charSet: CharSetUTF8,
			wantErr: false,
		},
		{
			name:    "UTF8 - Valid Multi-Byte Sequence",
			payload: "Hallo World 🌍",
			charSet: CharSetUTF8,
			wantErr: false,
		},
		{
			name:    "UTF8 - Valid Raw High-Byte",
			payload: "test \xFF byte",
			charSet: CharSetUTF8,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPacket(CommandPacket, tt.payload)

			_, err := Marshal(p, tt.charSet)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUnmarshal_CharSets(t *testing.T) {
	// helper function to create known byte slices
	makeRawData := func(payload []byte) []byte {
		length := uint32(len(payload) + 10)
		data := make([]byte, 12+len(payload)+2)

		// Length
		data[0] = byte(length)
		data[1] = byte(length >> 8)
		data[2] = byte(length >> 16)
		data[3] = byte(length >> 24)

		// ID: 1, Kind: 2
		data[4] = 1
		data[8] = 2

		// copy Payload
		copy(data[12:], payload)
		// Null-terminator already at the end as 0

		return data
	}

	tests := []struct {
		name     string
		rawBytes []byte
		charSet  CharSet
		wantErr  bool
	}{
		// --- CharSetASCII ---
		{
			name:     "Unmarshal ASCII - Valid",
			rawBytes: makeRawData([]byte("Status OK")),
			charSet:  CharSetASCII,
			wantErr:  false,
		},
		{
			name:     "Unmarshal ASCII - Invalid Byte (>127)",
			rawBytes: makeRawData([]byte{'S', 't', 'a', 't', 0x80}),
			charSet:  CharSetASCII,
			wantErr:  true,
		},

		// --- CharSetLatin_1 ---
		{
			name:     "Unmarshal Latin1 - Valid Extended Byte",
			rawBytes: makeRawData([]byte{'H', 0xE4, 'l', 'l', 'o'}), // 0xE4 (ä)
			charSet:  CharSetLatin1,
			wantErr:  false,
		},

		// --- CharSetUTF8 ---
		{
			name:     "Unmarshal UTF8 - Valid Any Byte",
			rawBytes: makeRawData([]byte{'H', 0xFF, 0xFE, 'o'}),
			charSet:  CharSetUTF8,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Packet
			err := Unmarshal(tt.rawBytes, &p, tt.charSet)

			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalAndUnmarshalRoundTrip(t *testing.T) {
	original := NewPacket(CommandPacket, "exec status")

	encoded, err := Marshal(original, CharSetUTF8)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded Packet
	err = Unmarshal(encoded, &decoded, CharSetUTF8)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if original.Length != decoded.Length {
		t.Errorf("Length mismatch: got %d, want %d", decoded.Length, original.Length)
	}
	if original.ID != decoded.ID {
		t.Errorf("ID mismatch: got %d, want %d", decoded.ID, original.ID)
	}
	if original.Kind != decoded.Kind {
		t.Errorf("Kind mismatch: got %d, want %d", decoded.Kind, original.Kind)
	}
	if original.Payload != decoded.Payload {
		t.Errorf("Payload mismatch: got %s, want %s", decoded.Payload, original.Payload)
	}
}
