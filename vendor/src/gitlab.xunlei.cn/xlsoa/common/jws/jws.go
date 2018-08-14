// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jws provides a partial implementation
// of JSON Web Signature encoding and decoding.
// It exists to support the golang.org/x/oauth2 package.
//
// See RFC 7515.
//
// Deprecated: this package is not intended for public use and might be
// removed in the future. It exists for internal use only.
// Please switch to another JWS package or copy this package into your own
// source tree.
package jws

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"crypto/hmac"
	"encoding/json"
	"bytes"
	"golang.org/x/oauth2/jws"
	"time"
)

func computeHmac(msg []byte, secret string) []byte {
	secretBytes := []byte(secret)
	h := hmac.New(sha256.New, secretBytes)
	h.Write(msg)
	return h.Sum(nil)
	//return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func Decode(payload string)(header *jws.Header, claimSet *jws.ClaimSet, headerBytes, claimSetBytes []byte){
	// decode returned id token to get expiry
	s := strings.Split(payload, ".")
	if len(s) < 2 {
		return nil, nil, nil, nil
	}
	var err error
	headerBytes, err = base64.RawURLEncoding.DecodeString(s[0])
	if err == nil {
		header = &jws.Header{}
		json.NewDecoder(bytes.NewBuffer(headerBytes)).Decode(header)
	}
	claimSetBytes, err = base64.RawURLEncoding.DecodeString(s[1])
	if err == nil {
		claimSet = &jws.ClaimSet{}
		err = json.NewDecoder(bytes.NewBuffer(claimSetBytes)).Decode(claimSet)
	}
	return header, claimSet, headerBytes, claimSetBytes
}

// Encode encodes a signed JWS with provided header and claim set.
func Encode(header *jws.Header, claimSet *jws.ClaimSet, secret string) (string, error){
	signer := func(data []byte)(sig []byte, err error){
		return computeHmac(data, secret), nil
	}
	return jws.EncodeWithSigner(header, claimSet, signer)
}

func IsClaimSetExpired(c *jws.ClaimSet) bool {
	return c.Exp <= time.Now().Unix()
}

// Verify tests whether the provided JWT token's signature was produced by the private key
// associated with the supplied public key.
func Verify(token string, secret string) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("jws: invalid token received, token must have 3 parts")
	}

	signedContent := parts[0] + "." + parts[1]
	signatureString := parts[2]
	hmacBytes := computeHmac([]byte(signedContent), secret)
	hmacString := base64.RawURLEncoding.EncodeToString(hmacBytes)
	if (signatureString != hmacString ) {
		return errors.New("jws: tonken is not pass")
	}
	return nil
}
