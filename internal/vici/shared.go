// this file contains the functions for managing shared secrets

package vici

import (
	"fmt"
)

type Key struct {
	ID     string   `json:"id,omitempty"`
	Typ    string   `json:"type"`
	Data   string   `json:"data"`
	Owners []string `json:"owners,omitempty"`
}

type UnloadKeyRequest struct {
	ID string `json:"id"`
}

type keyList struct {
	Keys []string `json:"keys"`
}

// load a shared secret into the IKE daemon
func (c *ClientConn) LoadShared(key *Key) error {
	msg, err := c.Request("load-shared", key)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("load-shared unsuccessful: %v", msg["errmsg"])
	}
	return nil
}

// unload (delete) a shared secret from the IKE daemon
func (c *ClientConn) UnloadShared(key *UnloadKeyRequest) error {
	msg, err := c.Request("unload-shared", key)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unload-shared unsuccessful: %v", msg["errmsg"])
	}
	return nil
}

// get a the names of the shared secrets currently loaded
func (c *ClientConn) GetShared() ([]string, error) {
	msg, err := c.Request("get-shared", nil)
	if err != nil {
		return nil, err
	}
	var keys keyList
	err = convertFromGeneral(msg, &keys)
	if err != nil {
		return nil, fmt.Errorf("convert response: %w", err)
	}
	return keys.Keys, nil
}
