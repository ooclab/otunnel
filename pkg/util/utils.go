package util

import (
	"crypto/sha1"

	"golang.org/x/crypto/pbkdf2"
)

var (
	salt = []byte("otunnel")
)

func GenSecret(secret string, keyiter int, keylen int) []byte {
	if keyiter == 0 {
		keyiter = int(secret[0])
	}
	if keylen == 0 {
		keylen = len(secret)
	}
	return pbkdf2.Key([]byte(secret), salt, keyiter, keylen, sha1.New)
}
