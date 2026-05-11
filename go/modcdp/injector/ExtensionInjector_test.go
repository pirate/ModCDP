package injector_test

import (
	"context"
	"encoding/json"
	"fmt"
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/injector"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type probeExtensionInjector struct {
	ExtensionInjector
}

func (i probeExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	return i.WaitForReadyServiceWorker(i.Options.InjectorServiceWorkerReadyTimeoutMS, true)
}

func TestExtensionInjectorProbesRealExtensionServiceWorkerWithSharedBaseConfig(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	nextID := 0
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		nextID++
		message := map[string]any{"id": nextID, "method": method, "params": params}
		if sessionID != "" {
			message["sessionId"] = sessionID
		}
		body, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		if err := wsutil.WriteClientText(conn, body); err != nil {
			return nil, err
		}
		for {
			raw, err := wsutil.ReadServerText(conn)
			if err != nil {
				return nil, err
			}
			var response map[string]any
			if err := json.Unmarshal(raw, &response); err != nil {
				return nil, err
			}
			responseID, _ := response["id"].(float64)
			if int(responseID) != nextID {
				continue
			}
			if errorObject, ok := response["error"].(map[string]any); ok {
				return nil, fmt.Errorf("%v", errorObject["message"])
			}
			result, _ := response["result"].(map[string]any)
			if result == nil {
				result = map[string]any{}
			}
			return result, nil
		}
	}

	injector := probeExtensionInjector{ExtensionInjector: NewExtensionInjector(ExtensionInjectorConfig{
		Send: send,
		AttachToTarget: func(targetID string) string {
			result, _ := send("Target.attachToTarget", map[string]any{"targetId": targetID, "flatten": true}, "")
			sessionID, _ := result["sessionId"].(string)
			return sessionID
		},
		InjectorExtensionID:              DefaultModCDPExtensionID,
		InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
		InjectorTrustServiceWorkerTarget: true,
	})}

	transportConfig := injector.GetTransportConfig()
	if transportConfig["injector_extension_id"] != DefaultModCDPExtensionID {
		t.Fatalf("injector_extension_id = %v", transportConfig["injector_extension_id"])
	}
	if len(injector.GetLauncherConfig().ExtraArgs) != 0 {
		t.Fatalf("expected empty launcher config")
	}
	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected service worker probe result")
	}
	if result.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", result.ExtensionID)
	}
	if filepath.Base(result.URL) != "service_worker.js" {
		t.Fatalf("URL = %q", result.URL)
	}
}

func TestExtensionInjectorKeepsModCDPServiceWorkerAliveThroughOffscreenKeepalive(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	nextID := 0
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		nextID++
		message := map[string]any{"id": nextID, "method": method, "params": params}
		if sessionID != "" {
			message["sessionId"] = sessionID
		}
		body, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		if err := wsutil.WriteClientText(conn, body); err != nil {
			return nil, err
		}
		for {
			raw, err := wsutil.ReadServerText(conn)
			if err != nil {
				return nil, err
			}
			var response map[string]any
			if err := json.Unmarshal(raw, &response); err != nil {
				return nil, err
			}
			responseID, _ := response["id"].(float64)
			if int(responseID) != nextID {
				continue
			}
			if errorObject, ok := response["error"].(map[string]any); ok {
				return nil, fmt.Errorf("%v", errorObject["message"])
			}
			result, _ := response["result"].(map[string]any)
			if result == nil {
				result = map[string]any{}
			}
			return result, nil
		}
	}

	injector := probeExtensionInjector{ExtensionInjector: NewExtensionInjector(ExtensionInjectorConfig{
		Send: send,
		AttachToTarget: func(targetID string) string {
			result, _ := send("Target.attachToTarget", map[string]any{"targetId": targetID, "flatten": true}, "")
			sessionID, _ := result["sessionId"].(string)
			return sessionID
		},
		InjectorExtensionID:              DefaultModCDPExtensionID,
		InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
		InjectorTrustServiceWorkerTarget: true,
	})}

	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected service worker probe result")
	}
	if result.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", result.ExtensionID)
	}

	foundOffscreen := false
	var contexts []any
	for attempt := 0; attempt < 50; attempt++ {
		contextsResult, err := send("Runtime.evaluate", map[string]any{
			"expression":    "chrome.runtime.getContexts({}).then((contexts) => contexts.map((context) => ({ type: context.contextType, url: context.documentUrl || context.origin || '' })))",
			"awaitPromise":  true,
			"returnByValue": true,
		}, result.SessionID)
		if err != nil {
			t.Fatal(err)
		}
		evaluation, _ := contextsResult["result"].(map[string]any)
		contexts, _ = evaluation["value"].([]any)
		for _, rawContext := range contexts {
			context, _ := rawContext.(map[string]any)
			if context["type"] == "OFFSCREEN_DOCUMENT" &&
				context["url"] == "chrome-extension://"+DefaultModCDPExtensionID+"/offscreen/keepalive.html" {
				foundOffscreen = true
			}
		}
		if foundOffscreen {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !foundOffscreen {
		t.Fatalf("expected offscreen keepalive context, got %#v", contexts)
	}

	time.Sleep(3 * time.Second)
	targetsResult, err := send("Target.getTargets", map[string]any{}, "")
	if err != nil {
		t.Fatal(err)
	}
	targets, _ := targetsResult["targetInfos"].([]any)
	foundServiceWorker := false
	for _, rawTarget := range targets {
		target, _ := rawTarget.(map[string]any)
		if target["type"] == "service_worker" &&
			target["url"] == "chrome-extension://"+DefaultModCDPExtensionID+"/modcdp/service_worker.js" {
			foundServiceWorker = true
		}
	}
	if !foundServiceWorker {
		t.Fatalf("expected service worker target, got %#v", targets)
	}
	version, err := send("Runtime.evaluate", map[string]any{
		"expression":    "globalThis.ModCDP?.__ModCDPServerVersion",
		"returnByValue": true,
	}, result.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	versionResult, _ := version["result"].(map[string]any)
	if versionResult["value"] != float64(1) {
		t.Fatalf("ModCDP server version = %#v", versionResult["value"])
	}
}

func TestExtensionInjectorOwnsSharedConfig(t *testing.T) {
	injector := NewExtensionInjector(ExtensionInjectorConfig{
		InjectorExtensionID:              "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
	})

	transportConfig := injector.GetTransportConfig()
	if transportConfig["injector_extension_id"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("injector_extension_id = %v", transportConfig["injector_extension_id"])
	}
	if len(injector.GetLauncherConfig().ExtraArgs) != 0 {
		t.Fatalf("expected empty launcher config")
	}
	if !injector.ServiceWorkerTargetMatches(map[string]any{
		"targetId": "target-1",
		"type":     "service_worker",
		"url":      "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/service_worker.js",
	}) {
		t.Fatalf("expected service worker target to match")
	}
	if _, err := injector.Inject(); err == nil {
		t.Fatalf("expected base Inject to fail")
	}
}

func TestExtensionInjectorSendWithTimeoutEnforcesCDPSendTimeout(t *testing.T) {
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	nextID := 0
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		nextID++
		message := map[string]any{"id": nextID, "method": method, "params": params}
		if sessionID != "" {
			message["sessionId"] = sessionID
		}
		body, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		if err := wsutil.WriteClientText(conn, body); err != nil {
			return nil, err
		}
		for {
			raw, err := wsutil.ReadServerText(conn)
			if err != nil {
				return nil, err
			}
			var response map[string]any
			if err := json.Unmarshal(raw, &response); err != nil {
				return nil, err
			}
			responseID, _ := response["id"].(float64)
			if int(responseID) != nextID {
				continue
			}
			if errorObject, ok := response["error"].(map[string]any); ok {
				return nil, fmt.Errorf("%v", errorObject["message"])
			}
			result, _ := response["result"].(map[string]any)
			if result == nil {
				result = map[string]any{}
			}
			return result, nil
		}
	}

	created, err := send("Target.createTarget", map[string]any{"url": "about:blank#modcdp-timeout"}, "")
	if err != nil {
		t.Fatal(err)
	}
	targetID, _ := created["targetId"].(string)
	attached, err := send("Target.attachToTarget", map[string]any{"targetId": targetID, "flatten": true}, "")
	if err != nil {
		t.Fatal(err)
	}
	sessionID, _ := attached["sessionId"].(string)
	if _, err := send("Runtime.enable", map[string]any{}, sessionID); err != nil {
		t.Fatal(err)
	}
	injector := NewExtensionInjector(ExtensionInjectorConfig{
		Send: send,
	})

	if _, err := injector.SendWithTimeout("Runtime.evaluate", map[string]any{
		"expression":   "new Promise(() => {})",
		"awaitPromise": true,
	}, sessionID, 5); err == nil || !strings.Contains(err.Error(), "Runtime.evaluate timed out after 5ms") {
		t.Fatalf("SendWithTimeout error = %v", err)
	}
}

func TestExtensionInjectorWakesConfiguredExtensionWithHiddenBackgroundTarget(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	nextID := 0
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		nextID++
		message := map[string]any{"id": nextID, "method": method, "params": params}
		if sessionID != "" {
			message["sessionId"] = sessionID
		}
		body, err := json.Marshal(message)
		if err != nil {
			return nil, err
		}
		if err := wsutil.WriteClientText(conn, body); err != nil {
			return nil, err
		}
		for {
			raw, err := wsutil.ReadServerText(conn)
			if err != nil {
				return nil, err
			}
			var response map[string]any
			if err := json.Unmarshal(raw, &response); err != nil {
				return nil, err
			}
			responseID, _ := response["id"].(float64)
			if int(responseID) != nextID {
				continue
			}
			if errorObject, ok := response["error"].(map[string]any); ok {
				return nil, fmt.Errorf("%v", errorObject["message"])
			}
			result, _ := response["result"].(map[string]any)
			if result == nil {
				result = map[string]any{}
			}
			return result, nil
		}
	}

	injector := NewExtensionInjector(ExtensionInjectorConfig{
		InjectorExtensionID: DefaultModCDPExtensionID,
		Send:                send,
	})

	if !injector.WakeConfiguredExtension() {
		t.Fatalf("expected wake to succeed")
	}
	targetsResult, err := send("Target.getTargets", map[string]any{}, "")
	if err != nil {
		t.Fatal(err)
	}
	targets, _ := targetsResult["targetInfos"].([]any)
	for _, rawTarget := range targets {
		target, _ := rawTarget.(map[string]any)
		if target["url"] == "chrome-extension://"+DefaultModCDPExtensionID+"/modcdp/wake.html" {
			return
		}
	}
	t.Fatalf("expected wake target, got %#v", targets)
}
