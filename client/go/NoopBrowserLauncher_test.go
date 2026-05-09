package modcdp

import "testing"

func TestNoopBrowserLauncherUsesNoBrowserLifecycleAndReturnsNoCDPEndpoints(t *testing.T) {
	launcher := NewNoopBrowserLauncher(LaunchOptions{
		CDPURL:      "ws://127.0.0.1:1/devtools/browser/not-used",
		UserDataDir: "/tmp/not-used-by-noop",
	})
	browser, err := launcher.Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if launcher.Launched != browser {
		t.Fatal("expected launcher to retain launched browser")
	}
	if browser.CDPURL != "" {
		t.Fatalf("CDPURL = %q", browser.CDPURL)
	}
	if browser.WSURL != "" {
		t.Fatalf("WSURL = %q", browser.WSURL)
	}
	if browser.PipeRead != nil || browser.PipeWrite != nil {
		t.Fatalf("expected no pipe handles")
	}
	browser.Close()
	browser.Close()
}
