package vici

import (
	"fmt"
)

func (c *ClientConn) ListConns(ike string) ([]map[string]IKEConf, error) {
	conns := []map[string]IKEConf{}
	var eventErr error
	var err error

	err = c.RegisterEvent("list-conn", func(response map[string]interface{}) {
		conn := &map[string]IKEConf{}
		err = ConvertFromGeneral(response, conn)
		if err != nil {
			eventErr = fmt.Errorf("list-conn event error: %w", err)
			return
		}
		conns = append(conns, *conn)
	})
	if err != nil {
		return nil, fmt.Errorf("error registering list-conn event: %w", err)
	}
	if eventErr != nil {
		return nil, eventErr
	}

	reqMap := map[string]interface{}{}
	if ike != "" {
		reqMap["ike"] = ike
	}

	_, err = c.Request("list-conns", reqMap)
	if err != nil {
		return nil, fmt.Errorf("error requesting list-conns: %w", err)
	}

	err = c.UnregisterEvent("list-conn")
	if err != nil {
		return nil, fmt.Errorf("error unregistering list-conns event: %w", err)
	}

	return conns, nil
}
