package modcdp

import "testing"

func TestRemoteBrowserLauncherRequiresUpstreamWSURLOrCDPURL(t *testing.T) {
	_, err := NewRemoteBrowserLauncher(LaunchOptions{}, "").Launch(LaunchOptions{})
	if err == nil || err.Error() != "launch.mode=remote requires upstream.ws_url or cdp_url" {
		t.Fatalf("Launch error = %v", err)
	}
}

func TestRemoteBrowserLauncherConnectsToRealBrowserFromHTTPAndWebSocketCDPEndpoints(t *testing.T) {
	local, err := NewLocalBrowserLauncher(LaunchOptions{}).Launch(LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer local.Close()

	httpLauncher := NewRemoteBrowserLauncher(LaunchOptions{}, local.CDPURL)
	fromHTTP, err := httpLauncher.Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if httpLauncher.Launched != fromHTTP {
		t.Fatal("expected launcher to retain launched browser")
	}
	httpTransportConfig := httpLauncher.GetTransportConfig()
	if httpTransportConfig["cdp_url"] != local.CDPURL {
		t.Fatalf("http transport cdp_url = %v, want %s", httpTransportConfig["cdp_url"], local.CDPURL)
	}
	if httpTransportConfig["ws_url"] != local.WSURL {
		t.Fatalf("http transport ws_url = %v, want %s", httpTransportConfig["ws_url"], local.WSURL)
	}
	if fromHTTP.CDPURL != local.CDPURL {
		t.Fatalf("fromHTTP.CDPURL = %q", fromHTTP.CDPURL)
	}
	if fromHTTP.WSURL != local.WSURL {
		t.Fatalf("fromHTTP.WSURL = %q, want %q", fromHTTP.WSURL, local.WSURL)
	}
	conn := connectBrowserbaseCDP(t, fromHTTP.WSURL)
	defer conn.Close()
	expectCDPBrowserSurface(t, conn)
	fromHTTP.Close()

	optionsLauncher := NewRemoteBrowserLauncher(LaunchOptions{CDPURL: local.CDPURL}, "")
	fromOptions, err := optionsLauncher.Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if fromOptions.CDPURL != local.CDPURL {
		t.Fatalf("fromOptions.CDPURL = %q", fromOptions.CDPURL)
	}
	if fromOptions.WSURL != local.WSURL {
		t.Fatalf("fromOptions.WSURL = %q, want %q", fromOptions.WSURL, local.WSURL)
	}
	fromOptions.Close()

	wsLauncher := NewRemoteBrowserLauncher(LaunchOptions{}, "")
	fromWS, err := wsLauncher.Launch(LaunchOptions{WSURL: local.WSURL})
	if err != nil {
		t.Fatal(err)
	}
	if wsLauncher.Launched != fromWS {
		t.Fatal("expected ws launcher to retain launched browser")
	}
	wsTransportConfig := wsLauncher.GetTransportConfig()
	if wsTransportConfig["cdp_url"] != local.WSURL {
		t.Fatalf("ws transport cdp_url = %v, want %s", wsTransportConfig["cdp_url"], local.WSURL)
	}
	if wsTransportConfig["ws_url"] != local.WSURL {
		t.Fatalf("ws transport ws_url = %v, want %s", wsTransportConfig["ws_url"], local.WSURL)
	}
	if fromWS.CDPURL != local.WSURL {
		t.Fatalf("fromWS.CDPURL = %q", fromWS.CDPURL)
	}
	if fromWS.WSURL != local.WSURL {
		t.Fatalf("fromWS.WSURL = %q", fromWS.WSURL)
	}
	expectCDPBrowserSurface(t, conn)
	fromWS.Close()
}
