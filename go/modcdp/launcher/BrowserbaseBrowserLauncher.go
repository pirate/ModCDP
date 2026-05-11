package launcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const DefaultBrowserbaseLauncherBaseURL = "https://api.browserbase.com"

var defaultBrowserbaseViewport = map[string]any{"width": 1288, "height": 711}

type BrowserbaseBrowserLauncher struct {
	BrowserLauncher
}

type browserbaseSession struct {
	ID                    string `json:"id"`
	ConnectURL            string `json:"connectUrl"`
	DebuggerURL           string `json:"debuggerUrl"`
	DebuggerFullscreenURL string `json:"debuggerFullscreenUrl"`
	Status                string `json:"status"`
}

func NewBrowserbaseBrowserLauncher(options LaunchOptions) *BrowserbaseBrowserLauncher {
	return &BrowserbaseBrowserLauncher{BrowserLauncher: NewBrowserLauncher(options)}
}

func (l *BrowserbaseBrowserLauncher) Launch(options LaunchOptions) (*LaunchedBrowser, error) {
	merged := mergeLaunchOptions(l.Options, options)
	browserbaseAPIKey := firstString(merged.BrowserbaseAPIKey, os.Getenv("BROWSERBASE_API_KEY"))
	if browserbaseAPIKey == "" {
		return nil, fmt.Errorf("launcher.launcher_mode=bb requires BROWSERBASE_API_KEY or launcher.launcher_options.browserbase_api_key")
	}

	baseURL := firstString(merged.BrowserbaseBaseURL, os.Getenv("BROWSERBASE_BASE_URL"), DefaultBrowserbaseLauncherBaseURL)
	resumeSessionID := firstString(merged.BrowserbaseSessionID)
	keepAlive := boolValue(merged.BrowserbaseKeepAlive, false)
	closeSessionOnClose := boolValue(merged.BrowserbaseCloseSessionOnClose, !keepAlive)

	createdSession := false
	var session browserbaseSession
	var err error
	if resumeSessionID != "" {
		err = browserbaseRequest(baseURL, browserbaseAPIKey, http.MethodGet, "/v1/sessions/"+resumeSessionID, nil, &session)
	} else {
		sessionCreateParams := objectValue(merged.BrowserbaseSessionCreateParams)
		browserSettings := mergeMap(objectValue(sessionCreateParams["browserSettings"]), objectValue(merged.BrowserbaseBrowserSettings))
		userMetadata := mergeMap(objectValue(sessionCreateParams["userMetadata"]), objectValue(merged.BrowserbaseUserMetadata))
		extensionID := firstString(
			merged.InjectorExtensionID,
			stringValue(sessionCreateParams["extensionId"]),
			stringValue(objectValue(sessionCreateParams["browserSettings"])["extensionId"]),
		)
		body := mergeMap(map[string]any{}, sessionCreateParams)
		if keepAlive {
			body["keepAlive"] = true
		}
		if region := firstString(merged.Region, stringValue(sessionCreateParams["region"])); region != "" {
			body["region"] = region
		}
		if merged.Timeout != 0 {
			body["timeout"] = merged.Timeout
		}
		if extensionID != "" {
			body["extensionId"] = extensionID
			browserSettings["extensionId"] = extensionID
		}
		if !hasViewportWidth(objectValue(browserSettings["viewport"])) {
			browserSettings["viewport"] = defaultBrowserbaseViewport
		}
		userMetadata["modcdp"] = "true"
		body["browserSettings"] = browserSettings
		body["userMetadata"] = userMetadata
		err = browserbaseRequest(baseURL, browserbaseAPIKey, http.MethodPost, "/v1/sessions", body, &session)
		createdSession = true
	}
	if err != nil {
		return nil, err
	}
	if session.ID == "" || session.ConnectURL == "" {
		return nil, fmt.Errorf("Browserbase session creation returned an unexpected shape")
	}

	closed := false
	close := func() {
		if closed {
			return
		}
		closed = true
		if !createdSession || !closeSessionOnClose {
			return
		}
		closeBrowserCDP(session.ConnectURL)
		body := map[string]any{"status": "REQUEST_RELEASE"}
		var ignored map[string]any
		_ = browserbaseRequest(baseURL, browserbaseAPIKey, http.MethodPost, "/v1/sessions/"+session.ID, body, &ignored)
	}

	launched := &LaunchedBrowser{
		// Browserbase ConnectURL is already a WebSocket CDP endpoint.
		CDPURL:                session.ConnectURL,
		BrowserbaseSessionID:  session.ID,
		BrowserbaseSessionURL: "https://www.browserbase.com/sessions/" + session.ID,
		BrowserbaseDebugURL:   firstString(session.DebuggerURL, session.DebuggerFullscreenURL),
		Close:                 close,
	}
	l.Launched = launched
	return launched, nil
}

func browserbaseRequest(baseURL string, browserbaseAPIKey string, method string, pathname string, body map[string]any, out any) error {
	var reader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(bodyBytes)
	}
	request, err := http.NewRequest(method, strings.TrimRight(baseURL, "/")+"/"+strings.TrimLeft(pathname, "/"), reader)
	if err != nil {
		return err
	}
	request.Header.Set("content-type", "application/json")
	request.Header.Set("x-bb-api-key", browserbaseAPIKey)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("Browserbase %s %s -> %d%s", method, pathname, response.StatusCode, errorText(responseBody))
	}
	if out == nil || len(responseBody) == 0 {
		return nil
	}
	return json.Unmarshal(responseBody, out)
}

func closeBrowserCDP(wsURL string) {
	if !strings.HasPrefix(wsURL, "ws://") && !strings.HasPrefix(wsURL, "wss://") {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, wsURL)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = wsutil.WriteClientMessage(conn, ws.OpText, []byte(`{"id":1,"method":"Browser.close","params":{}}`))
	_ = conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, _ = wsutil.ReadServerData(conn)
}

func errorText(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return ": " + string(body)
}

func boolValue(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func firstMap(values ...map[string]any) map[string]any {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func objectValue(value any) map[string]any {
	if object, ok := value.(map[string]any); ok {
		return object
	}
	return map[string]any{}
}

func mergeMap(left map[string]any, right map[string]any) map[string]any {
	merged := map[string]any{}
	for key, value := range left {
		merged[key] = value
	}
	for key, value := range right {
		merged[key] = value
	}
	return merged
}

func stringValue(value any) string {
	if value, ok := value.(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func hasViewportWidth(viewport map[string]any) bool {
	if _, ok := viewport["width"]; ok {
		return true
	}
	return false
}
