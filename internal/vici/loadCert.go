package vici

import (
	"fmt"
)

type certPayload struct {
	Typ  string `json:"type"` // (X509|X509_AC|X509_CRL)
	Flag string `json:"flag"` // (CA|AA|OCSP|NONE)
	Data string `json:"data"`
}

func (c *ClientConn) LoadCertificate(s string, typ string, flag string) error {
	msg, err := c.Request("load-cert", certPayload{
		Typ:  typ,
		Flag: flag,
		Data: s,
	})
	if err != nil {
		return fmt.Errorf("unsuccessful loadCert: %w", err)
	}

	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful loadCert: %v", msg["errmsg"])
	}

	return nil
}
