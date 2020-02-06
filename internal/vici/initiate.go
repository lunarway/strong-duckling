package vici

import (
	"fmt"
)

// Initiate is used to initiate an SA. This is the
// equivalent of `swanctl --initiate -c childname`
func (c *ClientConn) Initiate(child string, ike string) (err error) {
	request := map[string]interface{}{}
	if child != "" {
		request["child"] = child
	}
	if ike != "" {
		request["ike"] = ike
	}
	msg, err := c.Request("initiate", request)
	if err != nil {
		return err
	}
	if msg["success"] != "yes" {
		return fmt.Errorf("initiate unsuccessful: %v", msg["errmsg"])
	}
	return nil
}
