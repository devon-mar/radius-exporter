package collector

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"sync/atomic"
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
	AcceptCount         atomic.Uint64
	RejectCount         atomic.Uint64
	ValidMsgAuthCount   atomic.Uint64
	InvalidMsgAuthCount atomic.Uint64
	started             bool
	pc                  net.PacketConn
	ps                  *radius.PacketServer
}

func NewTestServer(t *testing.T, username string, password string, secret string) *TestServer {
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
			ts.AcceptCount.Add(1)
			code = radius.CodeAccessAccept
		} else {
			ts.RejectCount.Add(1)
			code = radius.CodeAccessReject
		}

		if err = ts.ValidateMessageAuthenticator(r.Packet, secret); err != nil {
			t.Logf("message authenticator validation failed: %v", err)
			ts.InvalidMsgAuthCount.Add(1)
		} else {
			ts.ValidMsgAuthCount.Add(1)
		}

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
		return fmt.Errorf("failed to set zero MessageAuthenticator: %w", err)
	}

	// Get the raw packet bytes for HMAC calculation
	rawPacket, err := packet.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode packet: %w", err)
	}

	// Calculate HMAC-MD5 of the entire packet using the shared secret
	h := hmac.New(md5.New, []byte(sharedSecret))
	h.Write(rawPacket)
	computedMessageAuth := h.Sum(nil)

	// Restore original Message-Authenticator
	err = rfc2869.MessageAuthenticator_Set(packet, originalMessageAuth)
	if err != nil {
		return fmt.Errorf("failed to set MessageAuthenticator: %w", err)
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

	ts := NewTestServer(t, user, password, secret)
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

	if ac := ts.AcceptCount.Load(); ac != 1 {
		t.Errorf("expected 1 Access-Accept, got %d", ac)
	}
	if rc := ts.RejectCount.Load(); rc != 0 {
		t.Errorf("expected 0 Access-Reject got %d", rc)
	}

	if c := ts.InvalidMsgAuthCount.Load(); c != 0 {
		t.Errorf("expected 0 Invalid MessageAuthenticators got %d", c)
	}

	if c := ts.ValidMsgAuthCount.Load(); c != 1 {
		t.Errorf("expected 1 Valid MessageAuthenticators got %d", c)
	}
}
