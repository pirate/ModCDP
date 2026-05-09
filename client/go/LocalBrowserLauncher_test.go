package modcdp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestLocalBrowserLauncherLaunchesRealBrowserAndSpeaksCDP(t *testing.T) {
	headless := true
	sandbox := false
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless: &headless,
		Sandbox:  &sandbox,
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.WSURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if err := wsutil.WriteClientText(conn, []byte(`{"id":1,"method":"Browser.getVersion","params":{}}`)); err != nil {
		t.Fatal(err)
	}
	body, err := wsutil.ReadServerText(conn)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		ID     int `json:"id"`
		Result struct {
			Product         string `json:"product"`
			ProtocolVersion string `json:"protocolVersion"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != 1 {
		t.Fatalf("unexpected response id %d", response.ID)
	}
	if !strings.Contains(response.Result.Product, "Chrome") && !strings.Contains(response.Result.Product, "Chromium") {
		t.Fatalf("unexpected product %q", response.Result.Product)
	}
	if response.Result.ProtocolVersion == "" {
		t.Fatal("expected protocolVersion")
	}
}

func TestLocalBrowserLauncherLaunchesRealBrowserOverRemoteDebuggingPipe(t *testing.T) {
	headless := true
	sandbox := false
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless: &headless,
		Sandbox:  &sandbox,
	}).Launch(LaunchOptions{RemoteDebugging: "pipe"})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()
	if !strings.HasPrefix(chrome.CDPURL, "pipe://") {
		t.Fatalf("CDPURL = %q", chrome.CDPURL)
	}
	if chrome.WSURL != "" {
		t.Fatalf("WSURL = %q", chrome.WSURL)
	}
	if chrome.PipeRead == nil || chrome.PipeWrite == nil {
		t.Fatal("expected pipe handles")
	}
	if err := writePipeMessage(chrome.PipeWrite, map[string]any{"id": 10, "method": "Browser.getVersion", "params": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	response, err := readPipeMessage(chrome.PipeRead)
	if err != nil {
		t.Fatal(err)
	}
	if response["id"] != float64(10) {
		t.Fatalf("response id = %v", response["id"])
	}
	result, _ := response["result"].(map[string]any)
	product, _ := result["product"].(string)
	if !strings.Contains(product, "Chrome") && !strings.Contains(product, "Chromium") {
		t.Fatalf("product = %q", product)
	}
}
