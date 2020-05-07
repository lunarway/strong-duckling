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

	defer func() {
		unregisterErr := c.UnregisterEvent("control-log")
		if unregisterErr != nil {
			if err == nil {
				err = unregisterErr
				return
			}
			err = fmt.Errorf("unregister control-log failed: %v: %w", unregisterErr, err)
		}
	}()

	err = c.RegisterEvent("control-log", func(response map[string]interface{}) {
		logger(response)
	})
	if err != nil {
		return fmt.Errorf("error registering control-log event: %w", err)
	}

	msg, err := c.Request("initiate", request)
	if msg["success"] != "yes" {
		return fmt.Errorf("initiate unsuccessful: %v", msg["errmsg"])
	}
	return nil
}
