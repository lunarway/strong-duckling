package vici

import (
	"fmt"
)

type Authorities struct {
	AuthorityMapping map[string]*AuthorityMapping `json:"authorities"`
}

type AuthorityMapping struct {
	Cacert      string   `json:"cacert,omitempty"`
	File        string   `json:"file,omitempty"`
	Handle      string   `json:"handle,omitempty"`
	Slot        string   `json:"slot,omitempty"`
	Module      string   `json:"module,omitempty"`
	CertUriBase string   `json:"cert_uri_base,omitempty"`
	CrlUris     []string `json:"crl_uris,omitempty"`
	OcspUris    []string `json:"ocsp_uris,omitempty"`
}

func (c *ClientConn) LoadAuthority(auth Authorities) error {
	requestMap := map[string]interface{}{}

	err := ConvertToGeneral(auth.AuthorityMapping, &requestMap)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	msg, err := c.Request("load-authority", requestMap)
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
	reqMap := &map[string]interface{}{}
	err := ConvertToGeneral(r, reqMap)
	if err != nil {
		return err
	}
	msg, err := c.Request("unload-authority", *reqMap)
	if err != nil {
		return err
	}

	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful UnloadAuthority %s", msg["errmsg"])
	}

	return nil
}
