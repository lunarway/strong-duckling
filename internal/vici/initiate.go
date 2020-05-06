package vici

import (
	"fmt"
)

// Initiate is used to initiate an SA. This is the
// equivalent of `swanctl --initiate -c childname`
func (c *ClientConn) Initiate(child string, ike string, logger func(fields map[string]interface{})) (err error) {
	request := map[string]interface{}{}
	if child != "" {
		request["child"] = child
	}
	if ike != "" {
		request["ike"] = ike
	}

	err = c.RegisterEvent("control-log", func(response map[string]interface{}) {
		logger(response)
	})
	if err != nil {
		return fmt.Errorf("error registering control-log event: %w", err)
	}

	msg, err := c.Request("initiate", request)
	if err != nil {
		_ = c.UnregisterEvent("control-log")
		return err
	}
	if msg["success"] != "yes" {
		err = c.UnregisterEvent("control-log")
		if err != nil {
			return fmt.Errorf("error unregistering control-log event: %w", err)
		}
		return fmt.Errorf("initiate unsuccessful: %v", msg["errmsg"])
	}
	err = c.UnregisterEvent("control-log")
	if err != nil {
		return fmt.Errorf("error unregistering control-log event: %w", err)
	}
	return nil
}
