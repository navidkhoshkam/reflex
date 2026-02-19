package inbound

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"

	"github.com/xtls/xray-core/proxy/reflex"
)

func BenchmarkSessionWriteFrame(b *testing.B) {
	sess, err := createTestSession()
	if err != nil {
		b.Fatalf("failed to create session: %v", err)
	}

	payload := bytes.Repeat([]byte("a"), 1024)
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := sess.WriteFrame(io.Discard, FrameTypeData, payload); err != nil {
			b.Fatalf("write frame failed: %v", err)
		}
	}
}

func BenchmarkSessionWriteFrameWithMorphing(b *testing.B) {
	sess, err := createTestSession()
	if err != nil {
		b.Fatalf("failed to create session: %v", err)
	}

	profile := GetProfileByName("youtube")
	profile.SetNextDelay(0) // avoid sleep inside benchmark loop
	payload := bytes.Repeat([]byte("b"), 1200)

	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		profile.SetNextDelay(0)
		if err := sess.WriteFrameWithMorphing(io.Discard, FrameTypeData, payload, profile); err != nil {
			b.Fatalf("write frame with morphing failed: %v", err)
		}
	}
}

func BenchmarkSessionReadFrame(b *testing.B) {
	key := bytes.Repeat([]byte{0x11}, 32)
	writerSess, err := NewSession(key)
	if err != nil {
		b.Fatalf("failed to create writer session: %v", err)
	}

	payload := bytes.Repeat([]byte("c"), 1024)
	var frameBytes bytes.Buffer
	if err := writerSess.WriteFrame(&frameBytes, FrameTypeData, payload); err != nil {
		b.Fatalf("failed to prebuild frame: %v", err)
	}
	raw := frameBytes.Bytes()

	b.SetBytes(int64(len(payload)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		readerSess, err := NewSession(key)
		if err != nil {
			b.Fatalf("failed to create reader session: %v", err)
		}
		if _, err := readerSess.ReadFrame(bytes.NewReader(raw)); err != nil {
			b.Fatalf("read frame failed: %v", err)
		}
	}
}

func FuzzParseDestination(f *testing.F) {
	f.Add([]byte{0x01, 127, 0, 0, 1, 0, 80})
	f.Add([]byte{0x03, 3, 'a', 'b', 'c', 0, 80})
	f.Add([]byte{0xFF, 0, 0, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = parseDestination(data)
	})
}

func FuzzSessionReadFrame(f *testing.F) {
	f.Add([]byte{0x00, 0x00, FrameTypeData})
	f.Add([]byte{0x00, 0x01, 0xFF, 0x00})
	f.Add([]byte{})

	key := bytes.Repeat([]byte{0x22}, 32)
	f.Fuzz(func(t *testing.T, data []byte) {
		sess, err := NewSession(key)
		if err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
		_, _ = sess.ReadFrame(bytes.NewReader(data))
	})
}

func ExampleNewSession() {
	key := bytes.Repeat([]byte{0x33}, 32)
	sess, err := NewSession(key)
	fmt.Println(err == nil && sess != nil)
	// Output: true
}

func ExampleSession_WriteFrame() {
	key := bytes.Repeat([]byte{0x44}, 32)
	sess, err := NewSession(key)
	if err != nil {
		fmt.Println(false)
		return
	}

	var wire bytes.Buffer
	_ = sess.WriteFrame(&wire, FrameTypeData, []byte("hello"))
	frame, err := sess.ReadFrame(bytes.NewReader(wire.Bytes()))
	fmt.Println(err == nil && string(frame.Payload) == "hello")
	// Output: true
}

func TestIntegrationHandleDataWithDestinationStable(t *testing.T) {
	h := createTestHandler()
	ctx := newCoreContextForTests(t)
	sess, err := createTestSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	initial := []byte{0x01, 127, 0, 0, 1, 0x00, 0x50, 'x'}
	reader := bufio.NewReader(bytes.NewReader(nil))
	conn := &bufferConn{}
	user := h.clients[0]

	if err := h.handleData(ctx, initial, conn, &testDispatcher{}, sess, user, reader); err != nil {
		t.Fatalf("handleData failed: %v", err)
	}
}

func TestIntegrationHandleDataFrameTypesStable(t *testing.T) {
	h := createTestHandler()
	ctx := newCoreContextForTests(t)
	sess, err := createTestSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Stream contains DATA, PADDING, TIMING, CLOSE frames.
	var stream bytes.Buffer
	if err := sess.WriteFrame(&stream, FrameTypeData, []byte("payload")); err != nil {
		t.Fatalf("failed to write data frame: %v", err)
	}
	if err := sess.WriteFrame(&stream, FrameTypePadding, []byte{0x00, 0x20}); err != nil {
		t.Fatalf("failed to write padding frame: %v", err)
	}
	timing := make([]byte, 8)
	if err := sess.WriteFrame(&stream, FrameTypeTiming, timing); err != nil {
		t.Fatalf("failed to write timing frame: %v", err)
	}
	if err := sess.WriteFrame(&stream, FrameTypeClose, nil); err != nil {
		t.Fatalf("failed to write close frame: %v", err)
	}

	initial := []byte{0x01, 127, 0, 0, 1, 0x00, 0x50, 'z'}
	reader := bufio.NewReader(bytes.NewReader(stream.Bytes()))
	conn := &bufferConn{}
	user := h.clients[0]

	if err := h.handleData(ctx, initial, conn, &testDispatcher{}, sess, user, reader); err != nil {
		t.Fatalf("handleData frame type flow failed: %v", err)
	}
}

func TestIntegrationHandleFallbackCompleteStable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_, _ = io.Copy(io.Discard, c)
		_, _ = c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nOK"))
	}()

	hAny, err := New(context.Background(), &Config{
		Fallback: &reflex.Fallback{Dest: uint32(ln.Addr().(*net.TCPAddr).Port)},
	})
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}
	h := hAny.(*Handler)

	clientConn, serverConn := net.Pipe()
	defer serverConn.Close()
	go func() {
		defer clientConn.Close()
		_, _ = clientConn.Write([]byte("GET / HTTP/1.1\r\nHost: x\r\n\r\n"))
	}()

	if err := h.handleFallback(newCoreContextForTests(t), bufio.NewReader(serverConn), serverConn); err != nil {
		msg := err.Error()
		if !strings.Contains(msg, "use of closed network connection") &&
			!strings.Contains(msg, "fallback connection ends") {
			t.Fatalf("unexpected fallback error: %v", err)
		}
	}
	<-done
}

func TestIntegrationHandleSessionEOFStable(t *testing.T) {
	h := createTestHandler()
	key := bytes.Repeat([]byte{0x55}, 32)
	conn := &bufferConn{}

	err := h.handleSession(
		context.Background(),
		bufio.NewReader(bytes.NewReader(nil)),
		conn,
		&testDispatcher{},
		key,
		h.clients[0],
		nil,
	)
	if err != nil {
		t.Fatalf("handleSession should return nil on EOF: %v", err)
	}
}

func TestIntegrationHandleSessionInvalidFrameStable(t *testing.T) {
	h := createTestHandler()
	key := bytes.Repeat([]byte{0x66}, 32)
	conn := &bufferConn{}

	// invalid frame type => ReadFrame error path in handleSession
	raw := []byte{0x00, 0x01, 0xFF, 0x00}
	err := h.handleSession(
		context.Background(),
		bufio.NewReader(bytes.NewReader(raw)),
		conn,
		&testDispatcher{},
		key,
		h.clients[0],
		nil,
	)
	if err == nil {
		t.Fatal("expected error for invalid frame")
	}
}


