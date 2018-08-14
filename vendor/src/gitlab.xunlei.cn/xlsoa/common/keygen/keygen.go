package keygen

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

const (
	maxSampleToRead = 20
)

type Generator struct {
	nano int64
}

func New() *Generator {
	return &Generator{nano: time.Now().UnixNano()}
}

func (g *Generator) AppId() (string, error) {
	return g.uniqueMD5Id()
}
func (g *Generator) ClientId() (string, string, error) {
	var id string
	var secret string
	var err error

	id, err = g.uniqueMD5Id()
	if err != nil {
		return "", "", err
	}

	secret, err = g.uniqueSHASecret()
	if err != nil {
		return "", "", err
	}

	return id, secret, nil
}

func (g *Generator) uniqueMD5Id() (string, error) {

	buf, err := g.sample()
	if err != nil {
		return "", err
	}

	out := fmt.Sprintf("%x", md5.Sum(buf))

	return out, nil
}

func (g *Generator) uniqueSHASecret() (string, error) {

	buf, err := g.sample()
	if err != nil {
		return "", err
	}

	s := sha256.Sum256(buf)
	out := base64.StdEncoding.EncodeToString(s[:])

	return out, nil
}

func (g *Generator) sample() ([]byte, error) {

	buf := make([]byte, maxSampleToRead)

	// Read sample from /dev/urandom
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	// Append UnixNano
	bb := bytes.NewBuffer(buf)
	binary.Write(bb, binary.BigEndian, g.nano)

	return bb.Bytes(), nil
}
