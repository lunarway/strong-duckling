package vici

import (
	"fmt"
	"net"
	"time"
)

const (
	DefaultReadTimeout = 15 * time.Second
)

// This object is not thread safe.
// if you want concurrent, you need create more clients.
type ClientConn struct {
	conn          net.Conn
	responseChan  chan segment
	eventHandlers map[string]func(response map[string]interface{})

	// ReadTimeout specifies a time limit for requests made
	// by this client.
	ReadTimeout time.Duration
}

func (c *ClientConn) Close() error {
	close(c.responseChan)
	return c.conn.Close()
}

func NewClientConn(conn net.Conn) *ClientConn {
	client := &ClientConn{
		conn:          conn,
		responseChan:  make(chan segment, 2),
		eventHandlers: map[string]func(response map[string]interface{}){},
		ReadTimeout:   DefaultReadTimeout,
	}
	return client
}

// Listen listens for data on configured net.Conn. This method is blocking until
// ClientConn.Close() is called or an unrecoverable error occours.
func (c *ClientConn) Listen() error {
	for {
		outMsg, err := readSegment(c.conn)
		if err != nil {
			return fmt.Errorf("vici: read segment: %w", err)
		}
		switch outMsg.typ {
		case stCMD_RESPONSE, stEVENT_CONFIRM:
			c.responseChan <- outMsg
		case stEVENT:
			handler := c.eventHandlers[outMsg.name]
			if handler != nil {
				handler(outMsg.msg)
			}
		default:
			return fmt.Errorf("vici: unprocessable message type '%s': raw message: %+v", outMsg.typ, outMsg)
		}
	}
}

func (c *ClientConn) Request(apiname string, concretePayload interface{}) (map[string]interface{}, error) {
	var request map[string]interface{}
	if concretePayload != nil {
		err := convertToGeneral(concretePayload, &request)
		if err != nil {
			return nil, fmt.Errorf("convert to general payload: %w", err)
		}
	}
	err := writeSegment(c.conn, segment{
		typ:  stCMD_REQUEST,
		name: apiname,
		msg:  request,
	})
	if err != nil {
		return nil, fmt.Errorf("writing segment: %w", err)
	}

	outMsg, err := c.readResponse()
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if outMsg.typ != stCMD_RESPONSE {
		return nil, fmt.Errorf("[%s] response error %d", apiname, outMsg.typ)
	}
	return outMsg.msg, nil
}

func (c *ClientConn) readResponse() (segment, error) {
	select {
	case outMsg := <-c.responseChan:
		return outMsg, nil
	case <-time.After(c.ReadTimeout):
		return segment{}, fmt.Errorf("timeout waiting for message response")
	}
}

func (c *ClientConn) RegisterEvent(name string, handler func(response map[string]interface{})) error {
	if c.eventHandlers[name] != nil {
		return fmt.Errorf("only one registration per name possible")
	}
	c.eventHandlers[name] = handler
	err := writeSegment(c.conn, segment{
		typ:  stEVENT_REGISTER,
		name: name,
	})
	if err != nil {
		delete(c.eventHandlers, name)
		return fmt.Errorf("write segment: %w", err)
	}
	outMsg, err := c.readResponse()
	if err != nil {
		delete(c.eventHandlers, name)
		return fmt.Errorf("read response: %w", err)
	}

	if outMsg.typ != stEVENT_CONFIRM {
		delete(c.eventHandlers, name)
		return fmt.Errorf("[event %s] response error %d", name, outMsg.typ)
	}
	return nil
}

func (c *ClientConn) UnregisterEvent(name string) error {
	err := writeSegment(c.conn, segment{
		typ:  stEVENT_UNREGISTER,
		name: name,
	})
	if err != nil {
		return fmt.Errorf("write segment: %w", err)
	}
	outMsg, err := c.readResponse()
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if outMsg.typ != stEVENT_CONFIRM {
		return fmt.Errorf("[event %s] response error %d", name, outMsg.typ)
	}
	delete(c.eventHandlers, name)
	return nil
}
