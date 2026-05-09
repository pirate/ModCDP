package modcdp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
		return nil, fmt.Errorf("launch.mode=bb requires BROWSERBASE_API_KEY or launch.options.browserbase_api_key")
	}

	projectID := firstString(
		merged.ProjectID,
		merged.BrowserbaseProjectID,
		os.Getenv("BROWSERBASE_PROJECT_ID"),
		os.Getenv("BB_PROJECT_ID"),
	)
	baseURL := firstString(merged.BaseURL, merged.BrowserbaseBaseURL, os.Getenv("BROWSERBASE_BASE_URL"), DefaultBrowserbaseLauncherBaseURL)
	resumeSessionID := firstString(merged.SessionID, merged.ResumeSessionID)
	keepAlive := boolValue(merged.KeepAlive, false)
	closeSessionOnClose := boolValue(merged.CloseSessionOnClose, !keepAlive)

	createdSession := false
	var session browserbaseSession
	var err error
	if resumeSessionID != "" {
		err = browserbaseRequest(baseURL, browserbaseAPIKey, http.MethodGet, "/v1/sessions/"+resumeSessionID, nil, &session)
	} else {
		sessionCreateParams := objectValue(firstMap(merged.SessionCreateParams, merged.BrowserbaseSessionCreateParams))
		browserSettings := mergeMap(objectValue(sessionCreateParams["browserSettings"]), objectValue(merged.BrowserSettings))
		userMetadata := mergeMap(objectValue(sessionCreateParams["userMetadata"]), objectValue(merged.UserMetadata))
		extensionID := firstString(
			merged.ExtensionID,
			stringValue(sessionCreateParams["extensionId"]),
			stringValue(objectValue(sessionCreateParams["browserSettings"])["extensionId"]),
		)
		body := mergeMap(map[string]any{}, sessionCreateParams)
		if projectID != "" {
			body["projectId"] = projectID
		}
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
		body := map[string]any{"status": "REQUEST_RELEASE"}
		if projectID != "" {
			body["projectId"] = projectID
		}
		var ignored map[string]any
		_ = browserbaseRequest(baseURL, browserbaseAPIKey, http.MethodPost, "/v1/sessions/"+session.ID, body, &ignored)
	}

	launched := &LaunchedBrowser{
		CDPURL:                session.ConnectURL,
		WSURL:                 session.ConnectURL,
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
