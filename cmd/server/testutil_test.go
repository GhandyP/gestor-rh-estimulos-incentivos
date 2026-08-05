package main

import (
	"bytes"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// syncBuffer es un buffer de bytes seguro para escrituras concurrentes: el
// logger slog escribe desde la goroutine de run() mientras el test lee.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// discardLogger descarta los eventos estructurados de slog en tests que no
// inspeccionan los logs.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTextLogger escribe eventos estructurados de slog en un syncBuffer.
func newTextLogger(buf *syncBuffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, nil))
}

// mustListen abre un listener en un puerto libre.
func mustListen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return ln
}

// setAuthEnv fija credenciales deterministas para el arranque del servidor.
func setAuthEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OPERATOR_USER", "testop")
	t.Setenv("OPERATOR_PASSWORD", "test-pass-123")
	t.Setenv("SESSION_SECRET", "test-session-secret-0123456789abcdef")
	t.Setenv("COOKIE_SECURE", "false")
}

// waitForListening espera a que el log estructurado reporte el listener y
// devuelve la dirección 127.0.0.1:puerto realmente ligada.
func waitForListening(t *testing.T, log *syncBuffer) string {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		for _, line := range strings.Split(log.String(), "\n") {
			if !strings.Contains(line, "server listening") {
				continue
			}
			idx := strings.Index(line, "addr=")
			if idx < 0 {
				continue
			}
			rest := line[idx+len("addr="):]
			addr := strings.Fields(rest)[0]
			if addr != "" {
				return addr
			}
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("el servidor no reportó 'server listening' en el log")
	return ""
}

// waitForStatus espera a que el endpoint devuelva el status esperado y lo
// retorna (tolera el arranque en curso del servidor).
func waitForStatus(t *testing.T, url string, want int) int {
	t.Helper()
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			code := resp.StatusCode
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if code == want {
				return code
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s no devolvió %d dentro del plazo", url, want)
	return 0
}

// dialAddr intenta abrir una conexión TCP a la dirección dada.
func dialAddr(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, 300*time.Millisecond)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

// freeAddr entrega una dirección 127.0.0.1 con un puerto libre.
func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reservar puerto: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}
