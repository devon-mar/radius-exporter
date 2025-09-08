package collector

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/devon-mar/radius-exporter/config"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2869"
)

const secret = "5ecr3t"

type TestServer struct {
	Address             string
	AcceptCount         int
	RejectCount         int
	ValidMsgAuthCount   int
	InvalidMsgAuthCount int
	mtx                 sync.Mutex
	started             bool
	pc                  net.PacketConn
	ps                  *radius.PacketServer
}

func NewTestServer(username string, password string, secret string) *TestServer {
	ts := &TestServer{}

	var err error
	ts.pc, err = net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		if ts.pc, err = net.ListenPacket("udp6", "[::1]:0"); err != nil {
			panic(fmt.Sprintf("error creating listener for test radius server: %v", err))
		}
	}
	ts.Address = ts.pc.LocalAddr().String()

	handler := func(w radius.ResponseWriter, r *radius.Request) {
		reqUsername := rfc2865.UserName_GetString(r.Packet)
		reqPassword := rfc2865.UserPassword_GetString(r.Packet)

		var code radius.Code
		if reqUsername == username && reqPassword == password {
			ts.mtx.Lock()
			ts.AcceptCount++
			ts.mtx.Unlock()
			code = radius.CodeAccessAccept
		} else {
			ts.mtx.Lock()
			ts.RejectCount++
			ts.mtx.Unlock()
			code = radius.CodeAccessReject
		}

		ts.mtx.Lock()
		if err = ts.ValidateMessageAuthenticator(r.Packet, secret); err != nil {
			ts.InvalidMsgAuthCount++
		} else {
			ts.ValidMsgAuthCount++
		}
		ts.mtx.Unlock()

		w.Write(r.Response(code))
	}

	ts.ps = &radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}

	return ts
}

// ValidateMessageAuthenticator validates the Message-Authenticator attribute in a RADIUS packet
// as per RFC 3579. It returns nil if valid, or an error if invalid.
func (ts *TestServer) ValidateMessageAuthenticator(packet *radius.Packet, sharedSecret string) error {
	// Check if Message-Authenticator attribute exists
	messageAuth := rfc2869.MessageAuthenticator_Get(packet)

	// Store original Message-Authenticator and zero it out for HMAC calculation
	originalMessageAuth := make([]byte, len(messageAuth))
	copy(originalMessageAuth, messageAuth)
	err := rfc2869.MessageAuthenticator_Set(packet, make([]byte, 16)) // Set to 16 zero bytes
	if err != nil {
		return errors.New("failed to set zero MessageAuthenticator: " + err.Error())
	}

	// Get the raw packet bytes for HMAC calculation
	rawPacket, err := packet.Encode()
	if err != nil {
		return errors.New("failed to encode packet: " + err.Error())
	}

	// Calculate HMAC-MD5 of the entire packet using the shared secret
	h := hmac.New(md5.New, []byte(sharedSecret))
	h.Write(rawPacket)
	computedMessageAuth := h.Sum(nil)

	// Restore original Message-Authenticator
	err = rfc2869.MessageAuthenticator_Set(packet, originalMessageAuth)
	if err != nil {
		return errors.New("failed to set MessageAuthenticator: " + err.Error())
	}

	// Compare computed HMAC with received Message-Authenticator
	if !hmac.Equal(computedMessageAuth, originalMessageAuth) {
		return errors.New("Message-Authenticator validation failed")
	}

	return nil
}

func (ts *TestServer) Start() {
	ts.started = true
	go func() { ts.ps.Serve(ts.pc) }()
}

func (ts *TestServer) Close() {
	if ts.started {
		ts.started = false
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()
		ts.ps.Shutdown(ctx)
		ts.pc.Close()
	}
}

func TestProbeAccessAccept(t *testing.T) {
	user := "testUser"
	password := "testPassowrd"

	ts := NewTestServer(user, password, secret)
	ts.Start()
	defer ts.Close()

	m := config.Module{
		Username:        user,
		Password:        password,
		Secret:          []byte(secret),
		Timeout:         time.Second * 2,
		Retry:           0,
		MaxPacketErrors: 0,
		NasID:           "test",
	}

	c := NewCollector(&ts.Address, &m)
	err := c.probe()
	if err != nil {
		t.Errorf("error probing RADIUS server: %v", err)
	}

	ts.Close()

	if ts.AcceptCount != 1 {
		t.Errorf("expected 1 Access-Accept, got %d", ts.AcceptCount)
	}
	if ts.RejectCount != 0 {
		t.Errorf("expected 0 Access-Reject got %d", ts.RejectCount)
	}

	if ts.InvalidMsgAuthCount != 0 {
		t.Errorf("expected 0 Invalid MessageAuthenticators got %d", ts.InvalidMsgAuthCount)
	}

	if ts.ValidMsgAuthCount != 1 {
		t.Errorf("expected 1 Valid MessageAuthenticators got %d", ts.InvalidMsgAuthCount)
	}
}
