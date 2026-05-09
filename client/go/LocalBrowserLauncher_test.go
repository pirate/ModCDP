package modcdp

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestLocalBrowserLauncherLaunchesRealBrowserAndSpeaksCDP(t *testing.T) {
	headless := true
	sandbox := false
	profileDir := t.TempDir()
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	launcher := NewLocalBrowserLauncher(LaunchOptions{
		Headless:                  &headless,
		Sandbox:                   &sandbox,
		ChromeReadyTimeoutMS:      45_000,
		ChromeReadyPollIntervalMS: 50,
	})
	chrome, err := launcher.Launch(LaunchOptions{
		Port:        port,
		UserDataDir: profileDir,
		ExtraArgs:   []string{"--window-size=900,700"},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		chrome.Close()
		if _, err := os.Stat(profileDir); err != nil {
			t.Fatalf("expected explicit user data dir to remain after close: %v", err)
		}
	}()
	if launcher.Launched != chrome {
		t.Fatal("expected launcher to retain launched browser")
	}
	if chrome.CDPURL != "http://127.0.0.1:"+strconv.Itoa(port) {
		t.Fatalf("CDPURL = %q", chrome.CDPURL)
	}
	if chrome.ProfileDir != profileDir {
		t.Fatalf("ProfileDir = %q, want %q", chrome.ProfileDir, profileDir)
	}
	transportConfig := launcher.GetTransportConfig()
	if transportConfig["cdp_url"] != chrome.CDPURL {
		t.Fatalf("transport cdp_url = %v, want %s", transportConfig["cdp_url"], chrome.CDPURL)
	}
	if transportConfig["ws_url"] != chrome.WSURL {
		t.Fatalf("transport ws_url = %v, want %s", transportConfig["ws_url"], chrome.WSURL)
	}
	if transportConfig["user_data_dir"] != chrome.ProfileDir {
		t.Fatalf("transport user_data_dir = %v, want %s", transportConfig["user_data_dir"], chrome.ProfileDir)
	}

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
	launcher := NewLocalBrowserLauncher(LaunchOptions{
		Headless:             &headless,
		Sandbox:              &sandbox,
		ChromeReadyTimeoutMS: 45_000,
	})
	chrome, err := launcher.Launch(LaunchOptions{RemoteDebugging: "pipe"})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()
	if launcher.Launched != chrome {
		t.Fatal("expected launcher to retain launched browser")
	}
	transportConfig := launcher.GetTransportConfig()
	if transportConfig["cdp_url"] != chrome.CDPURL {
		t.Fatalf("transport cdp_url = %v, want %s", transportConfig["cdp_url"], chrome.CDPURL)
	}
	if transportConfig["ws_url"] != "" {
		t.Fatalf("transport ws_url = %v", transportConfig["ws_url"])
	}
	if transportConfig["pipe_read"] != chrome.PipeRead {
		t.Fatal("expected transport pipe_read to use launched pipe")
	}
	if transportConfig["pipe_write"] != chrome.PipeWrite {
		t.Fatal("expected transport pipe_write to use launched pipe")
	}
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

func TestLocalBrowserLauncherCleansExplicitUserDataDirWhenRequested(t *testing.T) {
	headless := true
	sandbox := false
	cleanupUserDataDir := true
	profileDir, err := os.MkdirTemp("", "modcdp-go-local-profile-")
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless:             &headless,
		Sandbox:              &sandbox,
		ChromeReadyTimeoutMS: 45_000,
	}).Launch(LaunchOptions{
		UserDataDir:        profileDir,
		CleanupUserDataDir: &cleanupUserDataDir,
	})
	if err != nil {
		_ = os.RemoveAll(profileDir)
		t.Fatal(err)
	}

	chrome.Close()

	if _, err := os.Stat(profileDir); !os.IsNotExist(err) {
		_ = os.RemoveAll(profileDir)
		t.Fatalf("expected explicit user data dir to be removed, got %v", err)
	}
}
