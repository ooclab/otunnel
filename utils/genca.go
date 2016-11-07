package utils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	mathRand "math/rand"
	"time"
)

// GenCA 创建 ca
func GenCA(country string, org string, unit string, length int) {
	s1 := mathRand.NewSource(time.Now().UnixNano())
	r1 := mathRand.New(s1)

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(r1.Int63n(1000000)),
		Subject: pkix.Name{
			Country:            []string{country},
			Organization:       []string{org},
			OrganizationalUnit: []string{unit},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv, _ := rsa.GenerateKey(rand.Reader, length)
	pub := &priv.PublicKey
	caBuf, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}
	caFile := "ca.pem"
	ioutil.WriteFile(caFile, caBuf, 0600)
	fmt.Println("write to", caFile)

	privFile := "ca.key"
	privBuf := x509.MarshalPKCS1PrivateKey(priv)
	ioutil.WriteFile(privFile, privBuf, 0600)
	fmt.Println("write to", privFile)
}
