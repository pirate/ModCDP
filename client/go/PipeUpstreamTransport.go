package modcdp

import (
	"os"
	"sync"
)

type PipeUpstreamTransport struct {
	UpstreamTransport
	URL       string
	PipeRead  *os.File
	PipeWrite *os.File
	writeMu   sync.Mutex
	closed    bool
}

func NewPipeUpstreamTransport() *PipeUpstreamTransport {
	return &PipeUpstreamTransport{}
}

func (t *PipeUpstreamTransport) SetPipes(pipeRead *os.File, pipeWrite *os.File, url string) {
	t.PipeRead = pipeRead
	t.PipeWrite = pipeWrite
	t.URL = url
}

func (t *PipeUpstreamTransport) GetLauncherConfig() LaunchOptions {
	return LaunchOptions{RemoteDebugging: "pipe"}
}

func (t *PipeUpstreamTransport) Connect() error {
	if t.PipeRead == nil || t.PipeWrite == nil {
		return unimplementedUpstream("pipe")
	}
	t.closed = false
	go t.readLoop()
	return nil
}

func (t *PipeUpstreamTransport) Send(message map[string]any) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return writePipeMessage(t.PipeWrite, message)
}

func (t *PipeUpstreamTransport) Close() error {
	t.closed = true
	if t.PipeRead != nil {
		_ = t.PipeRead.Close()
	}
	if t.PipeWrite != nil {
		_ = t.PipeWrite.Close()
	}
	return nil
}

func (t *PipeUpstreamTransport) readLoop() {
	for {
		message, err := readPipeMessage(t.PipeRead)
		if err != nil {
			if !t.closed {
				t.emitClose(err)
			}
			return
		}
		t.emitRecv(message)
	}
}
