package modcdp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type probeExtensionInjector struct {
	ExtensionInjector
}

func (i probeExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	return i.WaitForReadyServiceWorker(i.Options.ServiceWorkerReadyTimeoutMS, true)
}

func TestExtensionInjectorProbesRealExtensionServiceWorkerWithSharedBaseConfig(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	conn, _, _, err := ws.Dial(ctx, chrome.WSURL)
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
		ExtensionID:               DefaultModCDPExtensionID,
		ServiceWorkerURLSuffixes:  []string{"/modcdp/service_worker.js"},
		TrustMatchedServiceWorker: true,
	})}

	transportConfig := injector.GetTransportConfig()
	if transportConfig["extension_id"] != DefaultModCDPExtensionID {
		t.Fatalf("extension_id = %v", transportConfig["extension_id"])
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

func TestExtensionInjectorOwnsSharedConfigAndRuntimeTransportConfig(t *testing.T) {
	injector := NewExtensionInjector(ExtensionInjectorConfig{
		ExtensionID:              "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
		ReverseProxyURL:          "ws://127.0.0.1:29292",
		NATSURL:                  "ws://127.0.0.1:4223",
		NATSSubjectPrefix:        "modcdp.test",
	})
	injector.Update(ExtensionInjectorConfig{NativeHostName: "com.modcdp.bridge"})

	transportConfig := injector.GetTransportConfig()
	if transportConfig["extension_id"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("extension_id = %v", transportConfig["extension_id"])
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

	extensionPath := t.TempDir()
	if err := injector.WriteExtensionRuntimeConfig(extensionPath); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(extensionPath, "modcdp", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  \"native_host_name\": \"com.modcdp.bridge\",\n  \"nats_subject_prefix\": \"modcdp.test\",\n  \"nats_url\": \"ws://127.0.0.1:4223\",\n  \"reverse_proxy_url\": \"ws://127.0.0.1:29292\"\n}\n"
	if string(config) != expected {
		t.Fatalf("config.json = %s", config)
	}
	if _, err := injector.Inject(); err == nil {
		t.Fatalf("expected base Inject to fail")
	}
}
