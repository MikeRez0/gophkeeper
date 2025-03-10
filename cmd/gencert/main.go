package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// создаём шаблон сертификата
	cert := &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"GophCorp LLC"},
			Country:      []string{"RU"},
			CommonName:   "localhost:8888",
		},
		// ExtraExtensions: []pkix.Extension{
		// 	{
		// 		Id:       asn1.ObjectIdentifier{2, 5, 29, 17},
		// 		Critical: false,
		// 		Value:    []byte("URI:https://localhost"),
		// 	},
		// },
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		IsCA:        true,
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("generate key error: %w", err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("create certificate error: %w", err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return fmt.Errorf("error encoding certificate : %w", err)
	}

	fc, err := os.Create(os.Args[1])
	if err != nil {
		return fmt.Errorf("error creating cert file: %w", err)
	}
	defer func() { _ = fc.Close() }()
	_, err = fc.Write(certPEM.Bytes())
	if err != nil {
		return fmt.Errorf("error writing cert file: %w", err)
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return fmt.Errorf("error encoding private key : %w", err)
	}

	fk, err := os.Create(os.Args[2])
	if err != nil {
		return fmt.Errorf("error creating key file: %w", err)
	}
	defer func() { _ = fk.Close() }()

	_, err = fk.Write(privateKeyPEM.Bytes())
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
