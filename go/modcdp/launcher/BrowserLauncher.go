package launcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pirate/ModCDP/go/modcdp/types"
)

type LaunchOptions = types.LaunchOptions
type ExtensionInjectorConfig = types.ExtensionInjectorConfig

const DefaultChromeReadyTimeoutMS = 45_000
const DefaultChromeReadyPollIntervalMS = 100

func boolPtr(value bool) *bool {
	return &value
}

var writePipeMessage = WritePipeMessage
var readPipeMessage = ReadPipeMessage

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func websocketURLFor(endpoint string) (string, error) {
	return WebsocketURLFor(endpoint)
}

func WebsocketURLFor(endpoint string) (string, error) {
	if strings.HasPrefix(endpoint, "ws://") || strings.HasPrefix(endpoint, "wss://") {
		return endpoint, nil
	}
	httpEndpoint := endpoint
	if !strings.Contains(endpoint, "://") {
		httpEndpoint = "http://" + endpoint
	}
	resp, err := http.Get(httpEndpoint + "/json/version")
	if err != nil {
		return "", fmt.Errorf("GET /json/version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		parsed, parseErr := url.Parse(httpEndpoint)
		if parseErr != nil {
			return "", parseErr
		}
		if parsed.Scheme == "https" {
			parsed.Scheme = "wss"
		} else {
			parsed.Scheme = "ws"
		}
		parsed.Path = "/devtools/browser"
		parsed.RawQuery = ""
		parsed.Fragment = ""
		return parsed.String(), nil
	}
	body, _ := io.ReadAll(resp.Body)
	var version map[string]any
	if err := json.Unmarshal(body, &version); err != nil {
		return "", fmt.Errorf("parse /json/version: %w", err)
	}
	wsURL, _ := version["webSocketDebuggerUrl"].(string)
	if wsURL == "" {
		return "", fmt.Errorf("HTTP discovery for %s returned no webSocketDebuggerUrl", endpoint)
	}
	return wsURL, nil
}

type LaunchedBrowser struct {
	// CDPURL is the effective CDP endpoint for the selected transport; launchers resolve HTTP discovery endpoints to ws:// before returning when they can.
	CDPURL                string   `json:"cdp_url,omitempty"`
	Close                 func()   `json:"-"`
	ProfileDir            string   `json:"profile_dir,omitempty"`
	PipeRead              *os.File `json:"-"`
	PipeWrite             *os.File `json:"-"`
	BrowserbaseSessionID  string   `json:"browserbase_session_id,omitempty"`
	BrowserbaseSessionURL string   `json:"browserbase_session_url,omitempty"`
	BrowserbaseDebugURL   string   `json:"browserbase_debug_url,omitempty"`
}

type BrowserLauncher struct {
	Options  LaunchOptions
	Launched *LaunchedBrowser
}

func NewBrowserLauncher(options LaunchOptions) BrowserLauncher {
	return BrowserLauncher{Options: options}
}

func (l *BrowserLauncher) Update(config LaunchOptions) *BrowserLauncher {
	l.Options = mergeLaunchOptions(l.Options, config)
	return l
}

func (l BrowserLauncher) GetTransportConfig() map[string]any {
	return map[string]any{
		"cdp_url":       firstString(launchedCDPURL(l.Launched), l.Options.CDPURL),
		"user_data_dir": firstString(launchedProfileDir(l.Launched), l.Options.UserDataDir),
		"pipe_read":     launchedPipeRead(l.Launched),
		"pipe_write":    launchedPipeWrite(l.Launched),
	}
}

func (l BrowserLauncher) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{
		InjectorBrowserbaseAPIKey:  l.Options.BrowserbaseAPIKey,
		InjectorBrowserbaseBaseURL: l.Options.BrowserbaseBaseURL,
		InjectorExtensionID:        l.Options.InjectorExtensionID,
	}
}

func (l BrowserLauncher) Launch(options LaunchOptions) (*LaunchedBrowser, error) {
	return nil, fmt.Errorf("%T.Launch is not implemented", l)
}

func mergeLaunchOptions(existing LaunchOptions, incoming LaunchOptions) LaunchOptions {
	merged := existing
	if incoming.ExecutablePath != "" {
		merged.ExecutablePath = incoming.ExecutablePath
	}
	if incoming.Port != 0 {
		merged.Port = incoming.Port
	}
	if incoming.RemoteDebugging != "" {
		merged.RemoteDebugging = incoming.RemoteDebugging
	}
	if incoming.UserDataDir != "" {
		merged.UserDataDir = incoming.UserDataDir
	}
	if incoming.CleanupUserDataDir != nil {
		merged.CleanupUserDataDir = incoming.CleanupUserDataDir
	}
	if incoming.ChromeReadyTimeoutMS != 0 {
		merged.ChromeReadyTimeoutMS = incoming.ChromeReadyTimeoutMS
	}
	if incoming.ChromeReadyPollIntervalMS != 0 {
		merged.ChromeReadyPollIntervalMS = incoming.ChromeReadyPollIntervalMS
	}
	if incoming.Headless != nil {
		merged.Headless = incoming.Headless
	}
	if incoming.Sandbox != nil {
		merged.Sandbox = incoming.Sandbox
	}
	if len(incoming.Args) > 0 {
		merged.Args = mergeChromeArgs(existing.Args, incoming.Args)
	}
	if len(incoming.ExtraArgs) > 0 {
		merged.ExtraArgs = mergeChromeArgs(existing.ExtraArgs, incoming.ExtraArgs)
	}
	if incoming.CDPURL != "" {
		merged.CDPURL = incoming.CDPURL
	}
	if incoming.BrowserbaseAPIKey != "" {
		merged.BrowserbaseAPIKey = incoming.BrowserbaseAPIKey
	}
	if incoming.BrowserbaseBaseURL != "" {
		merged.BrowserbaseBaseURL = incoming.BrowserbaseBaseURL
	}
	if incoming.BrowserbaseSessionID != "" {
		merged.BrowserbaseSessionID = incoming.BrowserbaseSessionID
	}
	if incoming.BrowserbaseKeepAlive != nil {
		merged.BrowserbaseKeepAlive = incoming.BrowserbaseKeepAlive
	}
	if incoming.BrowserbaseCloseSessionOnClose != nil {
		merged.BrowserbaseCloseSessionOnClose = incoming.BrowserbaseCloseSessionOnClose
	}
	if incoming.Region != "" {
		merged.Region = incoming.Region
	}
	if incoming.Timeout != 0 {
		merged.Timeout = incoming.Timeout
	}
	if incoming.InjectorExtensionID != "" {
		merged.InjectorExtensionID = incoming.InjectorExtensionID
	}
	if incoming.BrowserbaseBrowserSettings != nil {
		merged.BrowserbaseBrowserSettings = incoming.BrowserbaseBrowserSettings
	}
	if incoming.BrowserbaseUserMetadata != nil {
		merged.BrowserbaseUserMetadata = incoming.BrowserbaseUserMetadata
	}
	if incoming.BrowserbaseSessionCreateParams != nil {
		merged.BrowserbaseSessionCreateParams = incoming.BrowserbaseSessionCreateParams
	}
	return merged
}

func mergeChromeArgs(existing []string, incoming []string) []string {
	args := append(append([]string{}, existing...), incoming...)
	loadExtensionPaths := []string{}
	merged := []string{}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "--load-extension=") {
			merged = append(merged, arg)
			continue
		}
		for _, extensionPath := range strings.Split(strings.TrimPrefix(arg, "--load-extension="), ",") {
			if extensionPath != "" && !containsString(loadExtensionPaths, extensionPath) {
				loadExtensionPaths = append(loadExtensionPaths, extensionPath)
			}
		}
	}
	if len(loadExtensionPaths) > 0 {
		loadExtensionArg := "--load-extension=" + strings.Join(loadExtensionPaths, ",")
		firstURLIndex := -1
		for index, arg := range merged {
			if !strings.HasPrefix(arg, "-") {
				firstURLIndex = index
				break
			}
		}
		if firstURLIndex == -1 {
			merged = append(merged, loadExtensionArg)
		} else {
			merged = append(merged[:firstURLIndex], append([]string{loadExtensionArg}, merged[firstURLIndex:]...)...)
		}
	}
	return merged
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func firstString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func launchedCDPURL(launched *LaunchedBrowser) string {
	if launched == nil {
		return ""
	}
	return launched.CDPURL
}

func launchedProfileDir(launched *LaunchedBrowser) string {
	if launched == nil {
		return ""
	}
	return launched.ProfileDir
}

func launchedPipeRead(launched *LaunchedBrowser) *os.File {
	if launched == nil {
		return nil
	}
	return launched.PipeRead
}

func launchedPipeWrite(launched *LaunchedBrowser) *os.File {
	if launched == nil {
		return nil
	}
	return launched.PipeWrite
}
