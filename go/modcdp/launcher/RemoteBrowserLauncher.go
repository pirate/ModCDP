package launcher

import "fmt"

type RemoteBrowserLauncher struct {
	BrowserLauncher
	CDPURL string
}

func NewRemoteBrowserLauncher(options LaunchOptions, cdpURL string) *RemoteBrowserLauncher {
	if cdpURL != "" {
		options.CDPURL = cdpURL
	}
	return &RemoteBrowserLauncher{BrowserLauncher: NewBrowserLauncher(options), CDPURL: cdpURL}
}

func (l *RemoteBrowserLauncher) Launch(options LaunchOptions) (*LaunchedBrowser, error) {
	cdpURL := firstString(l.CDPURL, options.CDPURL, l.Options.CDPURL)
	if cdpURL == "" {
		return nil, fmt.Errorf("launcher.launcher_mode=remote requires upstream.upstream_cdp_url")
	}
	resolvedCDPURL, err := websocketURLFor(cdpURL)
	if err != nil {
		return nil, err
	}
	// CDPURL is resolved here so downstream transports can dial it directly.
	launched := &LaunchedBrowser{CDPURL: resolvedCDPURL, Close: func() {}}
	l.Launched = launched
	return launched, nil
}
