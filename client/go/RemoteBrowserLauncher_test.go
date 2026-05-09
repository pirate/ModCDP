package modcdp

import "testing"

func TestRemoteBrowserLauncherConnectsToRealBrowserFromHTTPAndWebSocketCDPEndpoints(t *testing.T) {
	local, err := NewLocalBrowserLauncher(LaunchOptions{}).Launch(LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer local.Close()

	fromHTTP, err := NewRemoteBrowserLauncher(LaunchOptions{}, local.CDPURL).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
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

	fromWS, err := NewRemoteBrowserLauncher(LaunchOptions{}, "").Launch(LaunchOptions{WSURL: local.WSURL})
	if err != nil {
		t.Fatal(err)
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
