package engine

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/apernet/OpenGFW/analyzer"
	"github.com/apernet/OpenGFW/io"
)

type lifecyclePacketIO struct {
	io.PacketIO
	registered chan struct{}
}

func (p *lifecyclePacketIO) Register(context.Context, io.PacketCallback) error {
	close(p.registered)
	return nil
}

type lifecycleLogger struct {
	Logger
	started chan struct{}
	release chan struct{}
	stopped chan struct{}
}

func (l *lifecycleLogger) WorkerStart(int) {
	l.started <- struct{}{}
}

func (l *lifecycleLogger) WorkerStop(int) {
	<-l.release
	l.stopped <- struct{}{}
}

func TestEngineWaitsForWorkers(t *testing.T) {
	const workerCount = 2
	packetIO := &lifecyclePacketIO{registered: make(chan struct{})}
	logger := &lifecycleLogger{
		started: make(chan struct{}, workerCount),
		release: make(chan struct{}),
		stopped: make(chan struct{}, workerCount),
	}
	release := sync.OnceFunc(func() { close(logger.release) })
	t.Cleanup(release)
	en, err := NewEngine(Config{IO: packetIO, Logger: logger, Workers: workerCount})
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan error, 1)
	go func() { finished <- en.Run(ctx) }()

	select {
	case <-packetIO.registered:
	case <-time.After(5 * time.Second):
		t.Fatal("packet IO was not registered")
	}
	for range workerCount {
		select {
		case <-logger.started:
		case <-time.After(5 * time.Second):
			t.Fatal("worker did not start")
		}
	}

	cancel()
	select {
	case <-finished:
		t.Fatal("engine returned before workers stopped")
	case <-time.After(100 * time.Millisecond):
	}

	release()
	select {
	case err := <-finished:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("engine did not return after workers stopped")
	}
	if got := len(logger.stopped); got != workerCount {
		t.Fatalf("stopped workers = %d, want %d", got, workerCount)
	}
}

type quotaTCPStream struct {
	data    []byte
	closed  bool
	limited bool
}

func (s *quotaTCPStream) Feed(_, _, _ bool, _ int, data []byte) (*analyzer.PropUpdate, bool) {
	s.data = append(s.data, data...)
	return nil, false
}

func (s *quotaTCPStream) Close(limited bool) *analyzer.PropUpdate {
	s.closed = true
	s.limited = limited
	return nil
}

func TestTCPAnalyzerByteQuota(t *testing.T) {
	tests := []struct {
		name       string
		quota      int
		wantData   string
		wantRemain int
		wantDone   bool
	}{
		{name: "quota exhausted", quota: 3, wantData: "abc", wantDone: true},
		{name: "quota remaining", quota: 8, wantData: "abcdef", wantRemain: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream := &quotaTCPStream{}
			entry := &tcpStreamEntry{Stream: stream, HasLimit: true, Quota: tt.quota}
			_, _, done := new(tcpStream).feedEntry(entry, false, false, false, 0, []byte("abcdef"))
			if !bytes.Equal(stream.data, []byte(tt.wantData)) || entry.Quota != tt.wantRemain ||
				done != tt.wantDone || stream.closed != tt.wantDone || stream.limited != tt.wantDone {
				t.Fatalf("data=%q, quota=%d, done=%t, closed=%t, limited=%t", stream.data, entry.Quota, done, stream.closed, stream.limited)
			}
		})
	}
}
