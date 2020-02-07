package vici

import (
	"fmt"
)

type Pool struct {
	PoolMapping map[string]interface{} `json:"pools"`
}

type PoolMapping struct {
	Addrs              string   `json:"addrs"`
	DNS                []string `json:"dns,omitempty"`
	NBNS               []string `json:"nbns,omitempty"`
	ApplicationVersion []string `json:"application_version,omitempty"`
	InternalIPv6Prefix []string `json:"internal_ipv6_prefix,omitempty"`
}

func (c *ClientConn) LoadPool(ph Pool) error {
	msg, err := c.Request("load-pool", ph.PoolMapping)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("load-pool unsuccessful: %v", msg["errmsg"])
	}
	return nil
}

type UnloadPoolRequest struct {
	Name string `json:"name"`
}

func (c *ClientConn) UnloadPool(r *UnloadPoolRequest) error {
	msg, err := c.Request("unload-pool", r)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unload-pool unsuccessful: %v", msg["errmsg"])
	}
	return nil
}
