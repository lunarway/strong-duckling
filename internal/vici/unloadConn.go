package vici

import (
	"fmt"
)

type UnloadConnRequest struct {
	Name string `json:"name"`
}

func (c *ClientConn) UnloadConn(r *UnloadConnRequest) error {
	msg, err := c.Request("unload-conn", r)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("unload-conn unsuccessful: %v", msg["errmsg"])
	}
	return nil
}
