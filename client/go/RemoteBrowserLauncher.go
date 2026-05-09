package modcdp

import "fmt"

type RemoteBrowserLauncher struct {
	BrowserLauncher
	CDPURL string
}

func NewRemoteBrowserLauncher(options LaunchOptions, cdpURL string) RemoteBrowserLauncher {
	return RemoteBrowserLauncher{BrowserLauncher: NewBrowserLauncher(options), CDPURL: cdpURL}
}

func (l RemoteBrowserLauncher) Launch(options LaunchOptions) (*LaunchedBrowser, error) {
	cdpURL := firstString(l.CDPURL, options.WSURL, options.CDPURL, l.Options.WSURL, l.Options.CDPURL)
	if cdpURL == "" {
		return nil, fmt.Errorf("launch.mode=remote requires upstream.ws_url")
	}
	wsURL, err := websocketURLFor(cdpURL)
	if err != nil {
		return nil, err
	}
	return &LaunchedBrowser{CDPURL: cdpURL, WSURL: wsURL, Close: func() {}}, nil
}
