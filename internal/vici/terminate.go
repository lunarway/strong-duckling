package vici

import (
	"fmt"
)

type TerminateRequest struct {
	Child    string `json:"child,omitempty"`
	Ike      string `json:"ike,omitempty"`
	Child_id string `json:"child-id,omitempty"`
	Ike_id   string `json:"ike-id,omitempty"`
	Force    string `json:"force,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Loglevel string `json:"loglevel,omitempty"`
}

// To be simple, kill a client that is connecting to this server. A client is a sa.
//Terminates an SA while streaming control-log events.
func (c *ClientConn) Terminate(r *TerminateRequest) error {
	msg, err := c.Request("terminate", r)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("terminate unsuccessful: %v", msg["errmsg"])
	}
	return nil
}
