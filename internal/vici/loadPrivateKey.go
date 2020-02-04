package vici

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type keyPayload struct {
	Typ  string `json:"type"`
	Data string `json:"data"`
}

// LoadECDSAPrivateKey encodes a *ecdsa.PrivateKey as a PEM block before sending
// it to the Vici interface
func (c *ClientConn) LoadECDSAPrivateKey(key *ecdsa.PrivateKey) error {
	mk, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "ECDSA PRIVATE KEY",
		Bytes: mk,
	})

	return c.loadPrivateKey("ECDSA", string(pemData))
}

// LoadRSAPrivateKey encodes a *rsa.PrivateKey as a PEM block before sending
// it to the Vici interface
func (c *ClientConn) LoadRSAPrivateKey(key *rsa.PrivateKey) error {
	mk := x509.MarshalPKCS1PrivateKey(key)

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: mk,
	})

	return c.loadPrivateKey("RSA", string(pemData))
}

// loadPrivateKey expects typ to be (RSA|ECDSA) and a PEM encoded data as a
// string
func (c *ClientConn) loadPrivateKey(typ, data string) error {
	requestMap := &map[string]interface{}{}

	k := keyPayload{
		Typ:  typ,
		Data: data,
	}
	err := convertToGeneral(k, requestMap)
	if err != nil {
		return fmt.Errorf("convert to general: %w", err)
	}

	msg, err := c.Request("load-key", *requestMap)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful loadPrivateKey: %v", msg["success"])
	}

	return nil
}
