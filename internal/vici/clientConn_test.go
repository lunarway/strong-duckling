package vici

import (
	"errors"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestClientConn_Close_listen tests that Listen returns when Close is called.
func TestClientConn_Close_listen(t *testing.T) {
	conn, _ := net.Pipe()
	defer conn.Close()

	client := NewClientConn(conn)

	var wg sync.WaitGroup
	defer wg.Wait()

	// wg will never be released if Listen does not terminated leading to a test
	// timeout.
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := client.Listen()
		t.Logf("Listen err: %v", err)
	}()

	err := client.Close()
	require.NoError(t, err, "expected a Listen error")
}

// TestClientConn_Listen_connClosed tests that Listen returns an error if the
// net.Conn is closed on.
func TestClientConn_Listen_connClosed(t *testing.T) {
	// create a net.Conn that is closed right away to simulate a failed network
	// connection.
	conn, _ := net.Pipe()
	conn.Close()

	client := NewClientConn(conn)
	defer client.Close()

	err := client.Listen()

	t.Logf("Err: %v", err)

	require.Error(t, err, "expected a Listen error")

	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatalf("Error expected to be io.ErrClosedPipe but was: %v", err)
	}
}

// TestClientConn_Listen_unhandleableSegmentType tests that Listen returns an
// error if un unprocessable segment type is received.
func TestClientConn_Listen_unprocessableSegmentType(t *testing.T) {
	viciConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var wg sync.WaitGroup
	defer wg.Wait()

	// write EVENT_UNKNOWN response from vici daemon to fake un unprocessable
	// payload
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer viciConn.Close()

		message := []byte{
			0x0, 0x0, 0x0, 0x1, // length 1 (single byte)
			byte(stEVENT_UNKNOWN), // unprocessable but valid segment type
		}
		_, err := viciConn.Write(message)
		if err != nil {
			t.Logf("Failed to write vici message: %v", err)
			return
		}
		t.Logf("Wrote message to vici conn: %v", message)
	}()

	client := NewClientConn(clientConn)
	defer client.Close()

	err := client.Listen()

	t.Logf("Listen err: %v", err)

	require.Error(t, err, "Listen() didn't return error")

	if errors.Is(err, io.EOF) {
		t.Fatalf("Listen() returned io.EOF: this happens on net.Conn.Close() so the listener did not stop in time")
	}
}
