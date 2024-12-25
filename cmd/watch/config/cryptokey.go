package config

import (
	"encoding/base64"
	"errors"
)

type CryptoKey []byte

var (
	DefaultPSK = [...]byte{
		0xd4, 0xf1, 0xbb, 0x3a, 0x20, 0x29, 0x07, 0x59,
		0xf0, 0xbc, 0xff, 0xab, 0xcf, 0x4e, 0x69, 0x01}
	EventPSK = [...]byte{
		0x38, 0x4b, 0xbc, 0xc0, 0x1d, 0xc0, 0x22, 0xd1, 0x81, 0xbf, 0x36,
		0xb8, 0x61, 0x21, 0xe1, 0xfb, 0x96, 0xb7, 0x2e, 0x55, 0xbf, 0x74,
		0x22, 0x7e, 0x9d, 0x6a, 0xfb, 0x48, 0xd6, 0x4c, 0xb1, 0xa1}
)

var EmptyKeyErr = errors.New("Empty Key String")
var InvalidKeyErr = errors.New("Invalid Key String")

func (b *CryptoKey) UnmarshalText(text []byte) error {

	if len(text) == 0 {
		return EmptyKeyErr
	}

	decoded, err := base64.StdEncoding.DecodeString(string(text))

	if err != nil {
		return err
	}

	if len(decoded) == 0 {
		return EmptyKeyErr
	}
	if len(decoded) == 1 {
		// this is silly
		decoded = DefaultPSK[:]
	}

	*b = decoded
	return nil
}

func (b CryptoKey) MarshalText() ([]byte, error) {
	return []byte(base64.StdEncoding.EncodeToString(b)), nil
}
func (b CryptoKey) String() string {
	return base64.StdEncoding.Strict().EncodeToString(b)
}
