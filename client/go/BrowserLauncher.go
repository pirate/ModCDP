package modcdp

import (
	"fmt"
	"os"
	"strings"
)

type LaunchedBrowser struct {
	CDPURL                string
	WSURL                 string
	Close                 func()
	ProfileDir            string
	PipeRead              *os.File
	PipeWrite             *os.File
	BrowserbaseSessionID  string
	BrowserbaseSessionURL string
	BrowserbaseDebugURL   string
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
		"ws_url":        firstString(launchedWSURL(l.Launched), l.Options.WSURL),
		"user_data_dir": firstString(launchedProfileDir(l.Launched), l.Options.UserDataDir),
		"pipe_read":     launchedPipeRead(l.Launched),
		"pipe_write":    launchedPipeWrite(l.Launched),
	}
}

func (l BrowserLauncher) GetInjectorConfig() map[string]any {
	return map[string]any{
		"browserbase_api_key":  l.Options.BrowserbaseAPIKey,
		"base_url":             l.Options.BaseURL,
		"browserbase_base_url": l.Options.BrowserbaseBaseURL,
		"extension_id":         l.Options.ExtensionID,
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
	if incoming.Headless != nil {
		merged.Headless = incoming.Headless
	}
	if incoming.Sandbox != nil {
		merged.Sandbox = incoming.Sandbox
	}
	if len(incoming.ExtraArgs) > 0 {
		merged.ExtraArgs = mergeChromeArgs(existing.ExtraArgs, incoming.ExtraArgs)
	}
	if incoming.CDPURL != "" {
		merged.CDPURL = incoming.CDPURL
	}
	if incoming.WSURL != "" {
		merged.WSURL = incoming.WSURL
	}
	if incoming.BrowserbaseAPIKey != "" {
		merged.BrowserbaseAPIKey = incoming.BrowserbaseAPIKey
	}
	if incoming.ProjectID != "" {
		merged.ProjectID = incoming.ProjectID
	}
	if incoming.BrowserbaseProjectID != "" {
		merged.BrowserbaseProjectID = incoming.BrowserbaseProjectID
	}
	if incoming.BaseURL != "" {
		merged.BaseURL = incoming.BaseURL
	}
	if incoming.BrowserbaseBaseURL != "" {
		merged.BrowserbaseBaseURL = incoming.BrowserbaseBaseURL
	}
	if incoming.SessionID != "" {
		merged.SessionID = incoming.SessionID
	}
	if incoming.ResumeSessionID != "" {
		merged.ResumeSessionID = incoming.ResumeSessionID
	}
	if incoming.KeepAlive != nil {
		merged.KeepAlive = incoming.KeepAlive
	}
	if incoming.CloseSessionOnClose != nil {
		merged.CloseSessionOnClose = incoming.CloseSessionOnClose
	}
	if incoming.Region != "" {
		merged.Region = incoming.Region
	}
	if incoming.Timeout != 0 {
		merged.Timeout = incoming.Timeout
	}
	if incoming.ExtensionID != "" {
		merged.ExtensionID = incoming.ExtensionID
	}
	if incoming.BrowserSettings != nil {
		merged.BrowserSettings = incoming.BrowserSettings
	}
	if incoming.UserMetadata != nil {
		merged.UserMetadata = incoming.UserMetadata
	}
	if incoming.SessionCreateParams != nil {
		merged.SessionCreateParams = incoming.SessionCreateParams
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

func launchedWSURL(launched *LaunchedBrowser) string {
	if launched == nil {
		return ""
	}
	return launched.WSURL
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
