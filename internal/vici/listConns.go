package vici

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c *ClientConn) ListConns(ike string) ([]map[string]IKEConf, error) {
	conns := []map[string]IKEConf{}
	var eventErr error
	var err error

	err = c.RegisterEvent("list-conn", func(response map[string]interface{}) {
		conn, err := mapConnections(response)
		if err != nil {
			eventErr = fmt.Errorf("list-conn event error: %w", err)
			return
		}
		conns = append(conns, conn)
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

func mapConnections(response map[string]interface{}) (map[string]IKEConf, error) {
	// first parse sets all fields of the IKE conf except the dynamic AuthConf
	// sections
	conn := map[string]IKEConf{}
	err := convertFromGeneral(response, &conn)
	if err != nil {
		return nil, fmt.Errorf("map base conn: %w", err)
	}

	// parse auth sections individually as they are a located on the root of the
	// IKEConf type separated by their key prefixes local-* and remote-*

	rawIKEConf := make(map[string]map[string]json.RawMessage)
	err = convertFromGeneral(response, &rawIKEConf)
	if err != nil {
		return nil, fmt.Errorf("map auth sections: %w", err)
	}

	for connName, ikeConfField := range rawIKEConf {
		currentConn := conn[connName]
		for key, value := range ikeConfField {
			// contains a pointer to the AuthConf map that we want to assign a
			// deserialized AuthConf to
			var destinationMap *map[string]AuthConf

			// filter fields by key name and set the destination map if applicable
			switch {
			case strings.HasPrefix(key, "local-"):
				if currentConn.LocalAuthSection == nil {
					currentConn.LocalAuthSection = make(map[string]AuthConf)
				}
				destinationMap = &currentConn.LocalAuthSection
			case strings.HasPrefix(key, "remote-"):
				if currentConn.RemoteAuthSection == nil {
					currentConn.RemoteAuthSection = make(map[string]AuthConf)
				}
				destinationMap = &currentConn.RemoteAuthSection
			default:
				// key is not an auth section and can be ignored
				continue
			}
			var authSection AuthConf
			err := json.Unmarshal(value, &authSection)
			if err != nil {
				return nil, fmt.Errorf("unmarshal auth conf: %w", err)
			}
			// write the auth section into the configured destination map
			(*destinationMap)[key] = authSection
		}
		// update the conn map with the updated conn struct
		conn[connName] = currentConn
	}
	return conn, nil
}
