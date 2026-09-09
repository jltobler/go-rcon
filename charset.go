package rcon

import (
	"errors"
	"unicode"
)

// CharSet defines the character set used for RCON payloads.
type CharSet int

// CharSet constants define the supported character sets.
const (
	CharSetASCII CharSet = iota + 1
	CharSetLatin1
	CharSetUTF8
)

// Errors define the possible errors that can occur when validating a byte against a character set.
var (
	ErrNonASCII       = errors.New("payload contains non-ASCII characters")
	ErrNonLatin       = errors.New("payload contains non-ISO-8859-1 characters")
	ErrNonUTF8        = errors.New("payload contains non-UTF-8 characters")
	ErrInvalidCharSet = errors.New("invalid character set")
)

// ValidateByte validates a single byte against the character set.
func (cs CharSet) ValidateByte(b byte) error {
	switch cs {
	case CharSetASCII:
		if b > unicode.MaxASCII {
			return ErrNonASCII
		}
	case CharSetLatin1:
		if b > unicode.MaxLatin1 {
			return ErrNonLatin
		}
	case CharSetUTF8:
		return nil
	default:
		return ErrInvalidCharSet
	}
	return nil
}
