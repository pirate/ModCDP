package modcdp

type NoopBrowserLauncher struct {
	BrowserLauncher
}

func NewNoopBrowserLauncher(options LaunchOptions) *NoopBrowserLauncher {
	return &NoopBrowserLauncher{BrowserLauncher: NewBrowserLauncher(options)}
}

func (l *NoopBrowserLauncher) Launch(options LaunchOptions) (*LaunchedBrowser, error) {
	launched := &LaunchedBrowser{Close: func() {}}
	l.Launched = launched
	return launched, nil
}
