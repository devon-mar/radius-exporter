package collector

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/devon-mar/radius-exporter/config"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

type TestServer struct {
	Address     string
	AcceptCount int
	RejectCount int
	mtx         sync.Mutex
	started     bool
	pc          net.PacketConn
	ps          *radius.PacketServer
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

		w.Write(r.Response(code))
	}

	ts.ps = &radius.PacketServer{
		Handler:      radius.HandlerFunc(handler),
		SecretSource: radius.StaticSecretSource([]byte(secret)),
	}

	return ts
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
	secret := "5ecr3t"

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
}
