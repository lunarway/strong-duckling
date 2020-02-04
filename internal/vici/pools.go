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
	requestMap := map[string]interface{}{}
	err := convertToGeneral(ph.PoolMapping, &requestMap)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	msg, err := c.Request("load-pool", requestMap)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unsuccessful LoadPool: %v", msg["success"])
	}

	return nil
}

type UnloadPoolRequest struct {
	Name string `json:"name"`
}

func (c *ClientConn) UnloadPool(r *UnloadPoolRequest) error {
	reqMap := &map[string]interface{}{}
	err := convertToGeneral(r, reqMap)
	if err != nil {
		return err
	}
	msg, err := c.Request("unload-pool", *reqMap)
	if err != nil {
		return err
	}

	if msg["success"] != "yes" {
		return fmt.Errorf("[Unload-Pool] %s", msg["errmsg"])
	}

	return nil
}
