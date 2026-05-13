package client

import (
	"path/filepath"
	"testing"
	"time"
)

const getTargetsOverride = `
async (params) => {
  const [upstream, tabs] = await Promise.all([
    ModCDP.sendLoopback("Target.getTargets", params),
    chrome.tabs.query({}),
  ]);

  const tabIdByUrl = new Map();
  for (const tab of tabs) {
    for (const url of [tab.url, tab.pendingUrl].filter(Boolean)) {
      if (!tabIdByUrl.has(url)) tabIdByUrl.set(url, tab.id);
    }
  }

  return {
    ...upstream,
    targetInfos: (upstream.targetInfos || []).map(targetInfo => ({
      ...targetInfo,
      tabId: tabIdByUrl.get(targetInfo.url) ?? null,
    })),
  };
}
`

const tabIDFromTargetIDCommand = `
async ({ targetId }) => {
  const targets = await chrome.debugger.getTargets();
  const target = targets.find(target => target.id === targetId);
  if (target?.tabId != null) return { tabId: target.tabId };
  const tabs = await chrome.tabs.query({});
  const tab = tabs.find(tab => target?.url && (tab.url === target.url || tab.pendingUrl === target.url));
  return { tabId: tab?.id ?? null };
}
`

const addTabIDMiddleware = `
async (payload, next) => {
  const seen = new WeakSet();
  const visit = async value => {
    if (!value || typeof value !== "object" || seen.has(value)) return;
    seen.add(value);
    if (!Array.isArray(value) && typeof value.targetId === "string" && value.tabId == null) {
      const { tabId } = await cdp.send("Custom.tabIdFromTargetId", { targetId: value.targetId });
      if (tabId != null) value.tabId = tabId;
    }
    for (const child of Array.isArray(value) ? value : Object.values(value)) await visit(child);
  };
  await visit(payload);
  return next(payload);
}
`

func TestModCDPClientRoutedDefaultOverrides(t *testing.T) {
	headless := true
	sandbox := false
	extensionPath, err := filepath.Abs("../../../dist/extension")
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless:  &headless,
		Sandbox:   &sandbox,
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "remote"},
		Upstream: UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: chrome.CDPURL},
		Injector: InjectorConfig{
			InjectorMode:                     "auto",
			InjectorExtensionPath:            extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Client: ClientConfig{
			ClientRoutes: map[string]string{
				"Target.getTargets":         "service_worker",
				"Target.createTarget":       "service_worker",
				"Target.setDiscoverTargets": "service_worker",
			},
		},
		Server: &ServerConfig{
			ServerLoopbackCDPURL: chrome.CDPURL,
			ServerRoutes:         map[string]string{"*.*": "loopback_cdp"},
		},
	})
	defer chrome.Close()
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.CDPURL != chrome.CDPURL {
		t.Fatalf("CDPURL = %q, expected %q", cdp.CDPURL, chrome.CDPURL)
	}
	if cdp.Server.ServerLoopbackCDPURL != chrome.CDPURL {
		t.Fatalf("ServerLoopbackCDPURL = %q, expected %q", cdp.Server.ServerLoopbackCDPURL, chrome.CDPURL)
	}

	rawTargets, err := cdp.sendMessage("Target.getTargets", map[string]any{}, "")
	if err != nil {
		t.Fatal(err)
	}
	targetInfos := targetInfosFromResult(t, rawTargets)
	if len(targetInfos) == 0 {
		t.Fatal("expected raw Target.getTargets targetInfos")
	}
	for _, targetInfo := range targetInfos {
		if _, ok := targetInfo["tabId"]; ok {
			t.Fatalf("raw CDP TargetInfo should not already contain tabId: %#v", targetInfo)
		}
	}

	if _, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name:       "Custom.tabIdFromTargetId",
		Expression: tabIDFromTargetIDCommand,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := cdp.Mod.AddMiddleware(CustomMiddleware{
		Name:       "*",
		Phase:      "response",
		Expression: addTabIDMiddleware,
	}); err != nil {
		t.Fatal(err)
	}
	middlewareTargets, err := cdp.Send("Target.getTargets", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hasPageTargetWithTabID(targetInfosFromResult(t, middlewareTargets)) {
		t.Fatal("wildcard response middleware should add tabId next to targetId inside TargetInfo")
	}

	if _, err := cdp.Mod.AddMiddleware(CustomMiddleware{
		Name:       "*",
		Phase:      "event",
		Expression: addTabIDMiddleware,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name:       "Target.getTargets",
		Expression: getTargetsOverride,
	}); err != nil {
		t.Fatal(err)
	}

	enrichedTargets, err := cdp.Send("Target.getTargets", nil)
	if err != nil {
		t.Fatal(err)
	}
	enrichedTargetInfos := targetInfosFromResult(t, enrichedTargets)
	if len(enrichedTargetInfos) == 0 {
		t.Fatal("expected enriched Target.getTargets targetInfos")
	}
	for _, targetInfo := range enrichedTargetInfos {
		if _, ok := targetInfo["tabId"]; !ok {
			t.Fatalf("every routed TargetInfo should include tabId: %#v", targetInfo)
		}
	}
	if !hasPageTargetWithTabID(enrichedTargetInfos) {
		t.Fatal("expected at least one page target to be matched to a chrome.tabs tab id")
	}

	if _, err := cdp.Mod.AddCustomEvent(CustomEvent{Name: "Target.targetCreated"}); err != nil {
		t.Fatal(err)
	}
	transformedEvents := make(chan map[string]any, 8)
	cdp.On("Target.targetCreated", func(data any) {
		payload, _ := data.(map[string]any)
		targetInfo, _ := payload["targetInfo"].(map[string]any)
		if targetInfo != nil {
			if _, ok := targetInfo["tabId"]; ok {
				transformedEvents <- payload
			}
		}
	})

	_, _ = cdp.Target.SetDiscoverTargets(TargetSetDiscoverTargetsParams{Discover: false})
	_, _ = cdp.Target.SetDiscoverTargets(TargetSetDiscoverTargetsParams{Discover: true})
	_, _ = cdp.Target.GetTargets()

	var event map[string]any
	select {
	case event = <-transformedEvents:
	case <-time.After(2 * time.Second):
		createdTarget, err := cdp.Target.CreateTarget(TargetCreateTargetParams{URL: "about:blank#modcdp-target-created"})
		if err != nil {
			t.Fatal(err)
		}
		_, _ = cdp.Target.GetTargets()
		select {
		case event = <-transformedEvents:
			targetInfo, _ := event["targetInfo"].(map[string]any)
			if targetInfo["targetId"] != createdTarget.TargetID {
				t.Fatalf("expected transformed Target.targetCreated for %s, got %#v", createdTarget.TargetID, targetInfo)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("expected transformed Target.targetCreated for %s", createdTarget.TargetID)
		}
	}
	targetInfo, _ := event["targetInfo"].(map[string]any)
	if _, ok := targetInfo["tabId"]; !ok {
		t.Fatalf("transformed event targetInfo should include tabId: %#v", targetInfo)
	}
	_, _ = cdp.Target.SetDiscoverTargets(TargetSetDiscoverTargetsParams{Discover: false})
}

func targetInfosFromResult(t *testing.T, result any) []map[string]any {
	t.Helper()
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("result is %T: %#v", result, result)
	}
	rawTargetInfos, ok := resultMap["targetInfos"].([]any)
	if !ok {
		t.Fatalf("targetInfos is %T: %#v", resultMap["targetInfos"], resultMap["targetInfos"])
	}
	targetInfos := make([]map[string]any, 0, len(rawTargetInfos))
	for _, rawTargetInfo := range rawTargetInfos {
		targetInfo, ok := rawTargetInfo.(map[string]any)
		if !ok {
			t.Fatalf("targetInfo is %T: %#v", rawTargetInfo, rawTargetInfo)
		}
		targetInfos = append(targetInfos, targetInfo)
	}
	return targetInfos
}

func hasPageTargetWithTabID(targetInfos []map[string]any) bool {
	for _, targetInfo := range targetInfos {
		if targetInfo["type"] == "page" {
			if _, ok := targetInfo["tabId"].(float64); ok {
				return true
			}
			if _, ok := targetInfo["tabId"].(int); ok {
				return true
			}
		}
	}
	return false
}
