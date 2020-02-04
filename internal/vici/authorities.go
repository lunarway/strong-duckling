package vici

import (
	"fmt"
)

type Authorities struct {
	AuthorityMapping map[string]*AuthorityMapping `json:"authorities"`
}

type AuthorityMapping struct {
	CACert      string   `json:"cacert,omitempty"`
	File        string   `json:"file,omitempty"`
	Handle      string   `json:"handle,omitempty"`
	Slot        string   `json:"slot,omitempty"`
	Module      string   `json:"module,omitempty"`
	CertURIBase string   `json:"cert_uri_base,omitempty"`
	CRLURIs     []string `json:"crl_uris,omitempty"`
	OCSPURIs    []string `json:"ocsp_uris,omitempty"`
}

func (c *ClientConn) LoadAuthority(auth Authorities) error {
	msg, err := c.Request("load-authority", auth.AuthorityMapping)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful LoadAuthority: %v", msg["success"])
	}

	return nil
}

type UnloadAuthorityRequest struct {
	Name string `json:"name"`
}

func (c *ClientConn) UnloadAuthority(r *UnloadAuthorityRequest) error {
	msg, err := c.Request("unload-authority", r)
	if err != nil {
		return err
	}

	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful UnloadAuthority %s", msg["errmsg"])
	}

	return nil
}
