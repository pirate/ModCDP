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

func TestExtensionsLoadUnpackedInjectorExercisesRealCDPLoadUnpackedPath(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
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

	injector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{
		Send:          send,
		ExtensionPath: extensionPath,
	})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("result = %#v", result)
	}
	if injector.LastError == nil {
		t.Fatal("expected LastError from real Extensions.loadUnpacked attempt")
	}
	message := injector.LastError.Error()
	if !strings.Contains(message, "Method not available") &&
		!strings.Contains(message, "Method not found") &&
		!strings.Contains(message, "wasn't found") {
		t.Fatalf("LastError = %v", injector.LastError)
	}
}
