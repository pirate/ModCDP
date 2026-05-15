package injector

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/browserbase/modcdp/go/modcdp/types"
)

const DefaultModCDPExtensionID = "mdedooklbnfejodmnhmkdpkaedafkehf"
const DefaultCDPSendTimeoutMS = 10_000
const DefaultExecutionContextTimeoutMS = 10_000
const DefaultServiceWorkerProbeTimeoutMS = 10_000
const DefaultServiceWorkerReadyTimeoutMS = 60_000
const DefaultServiceWorkerPollIntervalMS = 100
const DefaultTargetSessionPollIntervalMS = 20

var DefaultModCDPServiceWorkerURLSuffixes = []string{"/modcdp/service_worker.js"}
var extIDFromURL = regexp.MustCompile(`^chrome-extension://([a-z]+)/`)

//go:embed extension.zip
var bundledExtensionZip []byte

const modcdpReadyExpression = `Boolean(globalThis.ModCDP?.__ModCDPServerVersion >= 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)`

type SendCDP = types.SendCDP
type SessionIDForTarget = types.SessionIDForTarget
type AttachToTarget = types.AttachToTarget
type WaitForExecutionContext = types.WaitForExecutionContext
type LaunchOptions = types.LaunchOptions
type ExtensionInjectorConfig = types.ExtensionInjectorConfig
type ExtensionInjectionResult = types.ExtensionInjectionResult

func boolPtr(value bool) *bool {
	return &value
}

type ExtensionInjector struct {
	Options           ExtensionInjectorConfig
	UnusableTargetIDs map[string]bool
	LastError         error
}

func NewExtensionInjector(options ExtensionInjectorConfig) ExtensionInjector {
	if options.InjectorCDPSendTimeoutMS == 0 {
		options.InjectorCDPSendTimeoutMS = DefaultCDPSendTimeoutMS
	}
	if options.InjectorExecutionContextTimeoutMS == 0 {
		options.InjectorExecutionContextTimeoutMS = DefaultExecutionContextTimeoutMS
	}
	if options.InjectorServiceWorkerProbeTimeoutMS == 0 {
		options.InjectorServiceWorkerProbeTimeoutMS = DefaultServiceWorkerProbeTimeoutMS
	}
	if options.InjectorServiceWorkerReadyTimeoutMS == 0 {
		options.InjectorServiceWorkerReadyTimeoutMS = DefaultServiceWorkerReadyTimeoutMS
	}
	if options.InjectorServiceWorkerPollIntervalMS == 0 {
		options.InjectorServiceWorkerPollIntervalMS = DefaultServiceWorkerPollIntervalMS
	}
	if options.InjectorTargetSessionPollIntervalMS == 0 {
		options.InjectorTargetSessionPollIntervalMS = DefaultTargetSessionPollIntervalMS
	}
	return ExtensionInjector{Options: options, UnusableTargetIDs: map[string]bool{}}
}

func (i *ExtensionInjector) Update(config ExtensionInjectorConfig) *ExtensionInjector {
	if config.Send != nil {
		i.Options.Send = config.Send
	}
	if config.SessionIDForTarget != nil {
		i.Options.SessionIDForTarget = config.SessionIDForTarget
	}
	if config.AttachToTarget != nil {
		i.Options.AttachToTarget = config.AttachToTarget
	}
	if config.WaitForExecutionContext != nil {
		i.Options.WaitForExecutionContext = config.WaitForExecutionContext
	}
	if config.InjectorExtensionPath != "" {
		i.Options.InjectorExtensionPath = config.InjectorExtensionPath
	}
	if config.InjectorExtensionID != "" {
		i.Options.InjectorExtensionID = config.InjectorExtensionID
	}
	if config.InjectorServiceWorkerURLIncludes != nil {
		i.Options.InjectorServiceWorkerURLIncludes = append([]string{}, config.InjectorServiceWorkerURLIncludes...)
	}
	if config.InjectorServiceWorkerURLSuffixes != nil {
		i.Options.InjectorServiceWorkerURLSuffixes = append([]string{}, config.InjectorServiceWorkerURLSuffixes...)
	}
	if config.InjectorTrustServiceWorkerTarget {
		i.Options.InjectorTrustServiceWorkerTarget = true
	}
	if config.InjectorRequireServiceWorkerTarget {
		i.Options.InjectorRequireServiceWorkerTarget = true
	}
	if config.InjectorServiceWorkerReadyExpression != "" {
		i.Options.InjectorServiceWorkerReadyExpression = config.InjectorServiceWorkerReadyExpression
	}
	if config.InjectorCDPSendTimeoutMS != 0 {
		i.Options.InjectorCDPSendTimeoutMS = config.InjectorCDPSendTimeoutMS
	}
	if config.InjectorExecutionContextTimeoutMS != 0 {
		i.Options.InjectorExecutionContextTimeoutMS = config.InjectorExecutionContextTimeoutMS
	}
	if config.InjectorServiceWorkerProbeTimeoutMS != 0 {
		i.Options.InjectorServiceWorkerProbeTimeoutMS = config.InjectorServiceWorkerProbeTimeoutMS
	}
	if config.InjectorServiceWorkerReadyTimeoutMS != 0 {
		i.Options.InjectorServiceWorkerReadyTimeoutMS = config.InjectorServiceWorkerReadyTimeoutMS
	}
	if config.InjectorServiceWorkerPollIntervalMS != 0 {
		i.Options.InjectorServiceWorkerPollIntervalMS = config.InjectorServiceWorkerPollIntervalMS
	}
	if config.InjectorTargetSessionPollIntervalMS != 0 {
		i.Options.InjectorTargetSessionPollIntervalMS = config.InjectorTargetSessionPollIntervalMS
	}
	if config.InjectorBrowserbaseAPIKey != "" {
		i.Options.InjectorBrowserbaseAPIKey = config.InjectorBrowserbaseAPIKey
	}
	if config.InjectorBrowserbaseBaseURL != "" {
		i.Options.InjectorBrowserbaseBaseURL = config.InjectorBrowserbaseBaseURL
	}
	if config.UpstreamNativeMessagingHostName != "" {
		i.Options.UpstreamNativeMessagingHostName = config.UpstreamNativeMessagingHostName
	}
	if config.UpstreamNATSURL != "" {
		i.Options.UpstreamNATSURL = config.UpstreamNATSURL
	}
	if config.UpstreamNATSSubjectPrefix != "" {
		i.Options.UpstreamNATSSubjectPrefix = config.UpstreamNATSSubjectPrefix
	}
	return i
}

func (i ExtensionInjector) GetInjectorConfig() ExtensionInjectorConfig {
	return i.Options
}

func (i ExtensionInjector) GetLauncherConfig() LaunchOptions {
	return LaunchOptions{}
}

func (i ExtensionInjector) GetTransportConfig() map[string]any {
	if i.Options.InjectorExtensionID == "" {
		return map[string]any{}
	}
	return map[string]any{"injector_extension_id": i.Options.InjectorExtensionID}
}

func (i *ExtensionInjector) Prepare() error {
	return nil
}

func (i *ExtensionInjector) Close() error {
	return nil
}

func (i *ExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	return nil, fmt.Errorf("%T.Inject is not implemented", i)
}

func (i ExtensionInjector) readyExpression() string {
	if i.Options.InjectorServiceWorkerReadyExpression == "" {
		return modcdpReadyExpression
	}
	return fmt.Sprintf("(%s) && Boolean(%s)", modcdpReadyExpression, i.Options.InjectorServiceWorkerReadyExpression)
}

func (i ExtensionInjector) sendWithTimeout(method string, params map[string]any, sessionID string, timeoutMS int) (map[string]any, error) {
	if i.Options.Send == nil {
		return nil, fmt.Errorf("%T requires a CDP send function", i)
	}
	if params == nil {
		params = map[string]any{}
	}
	if timeoutMS == 0 {
		timeoutMS = i.Options.InjectorCDPSendTimeoutMS
	}
	if timeoutMS <= 0 {
		return i.Options.Send(method, params, sessionID)
	}
	type sendResult struct {
		result map[string]any
		err    error
	}
	done := make(chan sendResult, 1)
	go func() {
		result, err := i.Options.Send(method, params, sessionID)
		done <- sendResult{result: result, err: err}
	}()
	select {
	case result := <-done:
		return result.result, result.err
	case <-time.After(time.Duration(timeoutMS) * time.Millisecond):
		return nil, fmt.Errorf("%s timed out after %dms", method, timeoutMS)
	}
}

func (i ExtensionInjector) SendWithTimeout(method string, params map[string]any, sessionID string, timeoutMS int) (map[string]any, error) {
	return i.sendWithTimeout(method, params, sessionID, timeoutMS)
}

func (i ExtensionInjector) sessionIDForTarget(targetID string, timeoutMS int) string {
	deadline := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)
	for {
		if i.Options.SessionIDForTarget != nil {
			if sessionID := i.Options.SessionIDForTarget(targetID); sessionID != "" {
				return sessionID
			}
		}
		if timeoutMS <= 0 || time.Now().After(deadline) {
			return ""
		}
		time.Sleep(time.Duration(i.Options.InjectorTargetSessionPollIntervalMS) * time.Millisecond)
	}
}

func (i ExtensionInjector) ensureSessionIDForTarget(targetID string, timeoutMS int, allowAttach bool) string {
	if i.Options.SessionIDForTarget != nil {
		if sessionID := i.Options.SessionIDForTarget(targetID); sessionID != "" {
			return sessionID
		}
	}
	if allowAttach && i.Options.AttachToTarget != nil {
		if sessionID := i.Options.AttachToTarget(targetID); sessionID != "" {
			return sessionID
		}
	}
	return i.sessionIDForTarget(targetID, timeoutMS)
}

func (i ExtensionInjector) targetInfos() ([]map[string]any, error) {
	result, err := i.sendWithTimeout("Target.getTargets", map[string]any{}, "", i.Options.InjectorCDPSendTimeoutMS)
	if err != nil {
		return nil, err
	}
	rawTargets, _ := result["targetInfos"].([]any)
	targets := make([]map[string]any, 0, len(rawTargets))
	for _, rawTarget := range rawTargets {
		target, _ := rawTarget.(map[string]any)
		if target != nil {
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func (i ExtensionInjector) probeTarget(target map[string]any, sessionTimeoutMS int, allowAttach bool) (*ExtensionInjectionResult, error) {
	targetID, _ := target["targetId"].(string)
	targetURL, _ := target["url"].(string)
	if targetID == "" || i.UnusableTargetIDs[targetID] {
		return nil, nil
	}
	sessionID := i.ensureSessionIDForTarget(targetID, sessionTimeoutMS, allowAttach)
	if sessionID == "" {
		return nil, nil
	}
	if _, err := i.sendWithTimeout("Runtime.enable", map[string]any{}, sessionID, i.Options.InjectorCDPSendTimeoutMS); err != nil {
		return nil, err
	}
	probe, err := i.sendWithTimeout("Runtime.evaluate", map[string]any{
		"expression":    i.readyExpression(),
		"returnByValue": true,
	}, sessionID, i.Options.InjectorCDPSendTimeoutMS)
	if err != nil {
		return nil, err
	}
	result, _ := probe["result"].(map[string]any)
	if ready, _ := result["value"].(bool); !ready {
		return nil, nil
	}
	extensionID := ""
	if m := extIDFromURL.FindStringSubmatch(targetURL); len(m) > 1 {
		extensionID = m[1]
	}
	return &ExtensionInjectionResult{
		Source:      "discovered",
		ExtensionID: extensionID,
		TargetID:    targetID,
		URL:         targetURL,
		SessionID:   sessionID,
	}, nil
}

func (i ExtensionInjector) discoverReadyServiceWorker(matchedOnly bool) (*ExtensionInjectionResult, error) {
	targets, err := i.targetInfos()
	if err != nil {
		return nil, err
	}
	if i.Options.InjectorTrustServiceWorkerTarget {
		for _, target := range targets {
			if !i.serviceWorkerTargetMatches(target) {
				continue
			}
			probed, err := i.probeTarget(target, i.Options.InjectorServiceWorkerProbeTimeoutMS, true)
			if err != nil {
				return nil, err
			}
			if probed != nil {
				probed.Source = "trusted"
				return probed, nil
			}
		}
	}
	if i.Options.InjectorTrustServiceWorkerTarget || matchedOnly {
		return nil, nil
	}
	for _, target := range targets {
		targetType, _ := target["type"].(string)
		targetURL, _ := target["url"].(string)
		if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
			continue
		}
		probed, err := i.probeTarget(target, i.Options.InjectorServiceWorkerProbeTimeoutMS, false)
		if err == nil && probed != nil {
			return probed, nil
		}
	}
	return nil, nil
}

func (i ExtensionInjector) waitForReadyServiceWorker(timeoutMS int, matchedOnly bool) (*ExtensionInjectionResult, error) {
	deadline := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)
	for time.Now().Before(deadline) {
		discovered, err := i.discoverReadyServiceWorker(matchedOnly)
		if err != nil || discovered != nil {
			return discovered, err
		}
		time.Sleep(time.Duration(i.Options.InjectorServiceWorkerPollIntervalMS) * time.Millisecond)
	}
	return nil, nil
}

func (i ExtensionInjector) WaitForReadyServiceWorker(timeoutMS int, matchedOnly bool) (*ExtensionInjectionResult, error) {
	return i.waitForReadyServiceWorker(timeoutMS, matchedOnly)
}

func (i ExtensionInjector) serviceWorkerTargetMatches(target map[string]any) bool {
	targetURL, _ := target["url"].(string)
	targetType, _ := target["type"].(string)
	if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
		return false
	}
	hasExtensionID := i.Options.InjectorExtensionID != ""
	if i.Options.InjectorExtensionID != "" && !strings.HasPrefix(targetURL, "chrome-extension://"+i.Options.InjectorExtensionID+"/") {
		return false
	}
	for _, part := range i.Options.InjectorServiceWorkerURLIncludes {
		if !strings.Contains(targetURL, part) {
			return false
		}
	}
	if len(i.Options.InjectorServiceWorkerURLSuffixes) > 0 {
		matched := false
		for _, suffix := range i.Options.InjectorServiceWorkerURLSuffixes {
			if strings.HasSuffix(targetURL, suffix) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return hasExtensionID || len(i.Options.InjectorServiceWorkerURLIncludes) > 0 || len(i.Options.InjectorServiceWorkerURLSuffixes) > 0
}

func (i ExtensionInjector) ServiceWorkerTargetMatches(target map[string]any) bool {
	return i.serviceWorkerTargetMatches(target)
}

func prepareUnpackedExtension(extensionPath string) (string, string, error) {
	if extensionPath == "" {
		dir, err := os.MkdirTemp("", "modcdp-extension-")
		if err != nil {
			return "", "", err
		}
		reader, err := zip.NewReader(bytes.NewReader(bundledExtensionZip), int64(len(bundledExtensionZip)))
		if err != nil {
			_ = os.RemoveAll(dir)
			return "", "", err
		}
		if err := extractZipFiles(reader.File, dir); err != nil {
			_ = os.RemoveAll(dir)
			return "", "", err
		}
		return extensionRoot(dir), dir, nil
	}
	if !strings.HasSuffix(extensionPath, ".zip") {
		dir, err := os.MkdirTemp("", "modcdp-extension-")
		if err != nil {
			return "", "", err
		}
		if err := copyDir(extensionPath, dir); err != nil {
			_ = os.RemoveAll(dir)
			return "", "", err
		}
		return dir, dir, nil
	}
	dir, err := os.MkdirTemp("", "modcdp-extension-")
	if err != nil {
		return "", "", err
	}
	reader, err := zip.OpenReader(extensionPath)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", "", err
	}
	defer reader.Close()
	if err := extractZipFiles(reader.File, dir); err != nil {
		_ = os.RemoveAll(dir)
		return "", "", err
	}
	return extensionRoot(dir), dir, nil
}

func extractZipFiles(files []*zip.File, dir string) error {
	root, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		targetPath, err := safeZipTarget(root, file.Name)
		if err != nil {
			return err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			_ = src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		srcErr := src.Close()
		dstErr := dst.Close()
		if copyErr != nil {
			return copyErr
		}
		if srcErr != nil {
			return srcErr
		}
		if dstErr != nil {
			return dstErr
		}
	}
	return nil
}

func safeZipTarget(root string, name string) (string, error) {
	cleanName := filepath.Clean(name)
	if filepath.IsAbs(cleanName) || cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip entry %q escapes extension extraction directory", name)
	}
	targetPath := filepath.Join(root, cleanName)
	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}
	if targetAbs != root && !strings.HasPrefix(targetAbs, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("zip entry %q escapes extension extraction directory", name)
	}
	return targetAbs, nil
}

func extensionRoot(unpackedPath string) string {
	if _, err := os.Stat(filepath.Join(unpackedPath, "manifest.json")); err == nil {
		return unpackedPath
	}
	nested := filepath.Join(unpackedPath, "extension")
	if _, err := os.Stat(filepath.Join(nested, "manifest.json")); err == nil {
		return nested
	}
	return unpackedPath
}

func extensionIDFromManifestKey(extensionPath string) (string, error) {
	manifestBytes, err := os.ReadFile(filepath.Join(extensionPath, "manifest.json"))
	if err != nil {
		return "", nil
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", err
	}
	key, _ := manifest["key"].(string)
	if strings.TrimSpace(key) == "" {
		return "", nil
	}
	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(keyBytes)
	alphabet := "abcdefghijklmnop"
	result := strings.Builder{}
	for _, value := range digest[:16] {
		result.WriteByte(alphabet[value>>4])
		result.WriteByte(alphabet[value&0x0f])
	}
	return result.String(), nil
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, relative)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		sourceFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer sourceFile.Close()
		targetFile, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()
		_, err = io.Copy(targetFile, sourceFile)
		return err
	})
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
