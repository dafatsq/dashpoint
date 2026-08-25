package main

import "testing"

func TestServerFiberConfigUsesBrowserSafeHeaderBuffer(t *testing.T) {
	cfg := serverFiberConfig(nil)
	if cfg.ReadBufferSize < 32*1024 {
		t.Fatalf("expected read buffer to allow browser cookie headers, got %d", cfg.ReadBufferSize)
	}
}
