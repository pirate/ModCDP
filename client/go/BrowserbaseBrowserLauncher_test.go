package modcdp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const liveBrowserbaseTimeout = 120 * time.Second

func TestBrowserbaseBrowserLauncherCreatesVerifiesResumesAndReleasesRealSession(t *testing.T) {
	if strings.TrimSpace(os.Getenv("BROWSERBASE_API_KEY")) == "" {
		t.Skip("BROWSERBASE_API_KEY is required for live Browserbase tests")
	}
	options := LaunchOptions{
		ProjectID: os.Getenv("BROWSERBASE_PROJECT_ID"),
		Timeout:   120,
		BrowserSettings: map[string]any{
			"viewport":      map[string]any{"width": 900, "height": 700},
			"recordSession": false,
		},
		UserMetadata: map[string]any{
			"modcdp_launcher_test": "BrowserbaseBrowserLauncher",
		},
	}
	if region := os.Getenv("BROWSERBASE_REGION"); region != "" {
		options.Region = region
	}
	launcher := NewBrowserbaseBrowserLauncher(options)
	browser, err := launcher.Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var resumed *LaunchedBrowser
	var conn io.ReadWriteCloser
	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
		if resumed != nil {
			resumed.Close()
		}
		browser.Close()
		browser.Close()
	}()

	if browser.BrowserbaseSessionID == "" {
		t.Fatal("expected browserbase session id")
	}
	if launcher.Launched != browser {
		t.Fatal("expected launcher to retain launched browser")
	}
	transportConfig := launcher.GetTransportConfig()
	if transportConfig["cdp_url"] != browser.CDPURL {
		t.Fatalf("transport cdp_url = %v, want %s", transportConfig["cdp_url"], browser.CDPURL)
	}
	if transportConfig["ws_url"] != browser.WSURL {
		t.Fatalf("transport ws_url = %v, want %s", transportConfig["ws_url"], browser.WSURL)
	}
	if !strings.Contains(browser.BrowserbaseSessionURL, browser.BrowserbaseSessionID) {
		t.Fatalf("browserbase session url = %q", browser.BrowserbaseSessionURL)
	}
	if !strings.HasPrefix(browser.WSURL, "wss://") {
		t.Fatalf("ws url = %q", browser.WSURL)
	}
	if browser.CDPURL != browser.WSURL {
		t.Fatalf("cdp url = %q, ws url = %q", browser.CDPURL, browser.WSURL)
	}
	conn = connectBrowserbaseCDP(t, browser.WSURL)
	expectCDPBrowserSurface(t, conn)

	retrieved := retrieveBrowserbaseSession(t, browser.BrowserbaseSessionID)
	if retrieved["id"] != browser.BrowserbaseSessionID {
		t.Fatalf("retrieved id = %v", retrieved["id"])
	}
	if retrieved["status"] != "RUNNING" {
		t.Fatalf("retrieved status = %v", retrieved["status"])
	}
	if projectID := os.Getenv("BROWSERBASE_PROJECT_ID"); projectID != "" && retrieved["projectId"] != projectID {
		t.Fatalf("retrieved projectId = %v", retrieved["projectId"])
	}

	closeSessionOnClose := false
	resumed, err = NewBrowserbaseBrowserLauncher(LaunchOptions{
		SessionID:           browser.BrowserbaseSessionID,
		CloseSessionOnClose: &closeSessionOnClose,
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resumed.BrowserbaseSessionID != browser.BrowserbaseSessionID {
		t.Fatalf("resumed session id = %q", resumed.BrowserbaseSessionID)
	}
	if !strings.HasPrefix(resumed.WSURL, "wss://") {
		t.Fatalf("resumed ws url = %q", resumed.WSURL)
	}
	expectCDPBrowserSurface(t, conn)

	_ = conn.Close()
	conn = nil
	resumed.Close()
	browser.Close()
	browser.Close()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if retrieveBrowserbaseSession(t, browser.BrowserbaseSessionID)["status"] != "RUNNING" {
			return
		}
		time.Sleep(time.Second)
	}
	t.Fatal("Browserbase session did not leave RUNNING status after release")
}

func connectBrowserbaseCDP(t *testing.T, rawURL string) io.ReadWriteCloser {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), liveBrowserbaseTimeout)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func expectCDPBrowserSurface(t *testing.T, conn io.ReadWriter) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"id": 1, "method": "Browser.getVersion", "params": map[string]any{}})
	if err := wsutil.WriteClientText(conn, body); err != nil {
		t.Fatal(err)
	}
	data, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		t.Fatal(err)
	}
	var message map[string]any
	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatal(err)
	}
	result, _ := message["result"].(map[string]any)
	if _, ok := result["product"].(string); !ok {
		t.Fatalf("Browser.getVersion result = %#v", message)
	}
}

func retrieveBrowserbaseSession(t *testing.T, sessionID string) map[string]any {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, browserbaseAPIURL("/v1/sessions/"+sessionID), nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("x-bb-api-key", os.Getenv("BROWSERBASE_API_KEY"))
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("Browserbase session fetch returned %d: %s", response.StatusCode, string(body))
	}
	var session map[string]any
	if err := json.Unmarshal(body, &session); err != nil {
		t.Fatal(err)
	}
	return session
}

func browserbaseAPIURL(pathname string) string {
	baseURL := firstString(os.Getenv("BROWSERBASE_BASE_URL"), DefaultBrowserbaseLauncherBaseURL)
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(pathname, "/")
}

func TestBrowserbaseBrowserLauncherRequiresAPIKey(t *testing.T) {
	if strings.TrimSpace(os.Getenv("BROWSERBASE_API_KEY")) != "" {
		t.Skip("BROWSERBASE_API_KEY is set")
	}
	_, err := NewBrowserbaseBrowserLauncher(LaunchOptions{}).Launch(LaunchOptions{})
	if err == nil || !strings.Contains(err.Error(), "BROWSERBASE_API_KEY") {
		t.Fatalf("expected missing key error, got %v", err)
	}
}

func ExampleBrowserbaseBrowserLauncher_options() {
	_ = LaunchOptions{
		ProjectID: "project-id",
		BrowserSettings: map[string]any{
			"viewport": map[string]any{"width": 900, "height": 700},
		},
		UserMetadata: map[string]any{"modcdp_launcher_test": "BrowserbaseBrowserLauncher"},
	}
	fmt.Println("ok")
	// Output: ok
}
