package transport

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const DefaultUpstreamNativeMessagingHostName = "com.modcdp.bridge"
const DefaultUpstreamNativeMessagingWaitTimeoutMS = 10_000

const nativeHostConfigEnv = "MODCDP_NATIVE_HOST_CONFIG"

func init() {
	if configPath := os.Getenv(nativeHostConfigEnv); configPath != "" {
		if err := runNativeMessagingHost(configPath); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

type NativeMessagingUpstreamTransport struct {
	UpstreamTransport
	UpstreamNativeMessagingManifest                string
	UpstreamNativeMessagingManifests               []string
	IncludeDefaultUpstreamNativeMessagingManifests bool
	UpstreamNativeMessagingHostName                string
	ExtensionID                                    string
	WaitTimeoutMS                                  int
	URL                                            string
	Listener                                       net.Listener
	Conn                                           net.Conn
	BoundPort                                      int
	CDPURL                                         string
	UserDataDir                                    string
	writeMu                                        sync.Mutex
	peerCh                                         chan struct{}
	peerOnce                                       sync.Once
	closeCh                                        chan struct{}
	stateMu                                        sync.Mutex
	closed                                         bool
}

type NativeMessagingUpstreamTransportOptions struct {
	UpstreamNativeMessagingManifest      string   `json:"upstream_nativemessaging_manifest,omitempty"`
	UpstreamNativeMessagingManifests     []string `json:"upstream_nativemessaging_manifests,omitempty"`
	UpstreamNativeMessagingHostName      string   `json:"upstream_nativemessaging_host_name,omitempty"`
	InjectorExtensionID                  string   `json:"injector_extension_id,omitempty"`
	UpstreamNativeMessagingWaitTimeoutMS int      `json:"upstream_nativemessaging_wait_timeout_ms,omitempty"`
}

func NewNativeMessagingUpstreamTransport(options NativeMessagingUpstreamTransportOptions) *NativeMessagingUpstreamTransport {
	nativeHostName := firstNonEmptyString(options.UpstreamNativeMessagingHostName, DefaultUpstreamNativeMessagingHostName)
	extensionID := firstNonEmptyString(options.InjectorExtensionID, DefaultModCDPExtensionID)
	waitTimeoutMS := options.UpstreamNativeMessagingWaitTimeoutMS
	if waitTimeoutMS == 0 {
		waitTimeoutMS = DefaultUpstreamNativeMessagingWaitTimeoutMS
	}
	return &NativeMessagingUpstreamTransport{
		UpstreamNativeMessagingManifest:                options.UpstreamNativeMessagingManifest,
		UpstreamNativeMessagingManifests:               append([]string{}, options.UpstreamNativeMessagingManifests...),
		IncludeDefaultUpstreamNativeMessagingManifests: options.UpstreamNativeMessagingManifest == "" && len(options.UpstreamNativeMessagingManifests) == 0,
		UpstreamNativeMessagingHostName:                nativeHostName,
		ExtensionID:                                    extensionID,
		WaitTimeoutMS:                                  waitTimeoutMS,
		peerCh:                                         make(chan struct{}),
		closeCh:                                        make(chan struct{}),
	}
}

func (t *NativeMessagingUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	shouldInstallNativeHost := false
	if value, ok := config["upstream_nativemessaging_manifest"]; ok {
		t.UpstreamNativeMessagingManifest, _ = value.(string)
		shouldInstallNativeHost = true
	}
	if value, ok := config["upstream_nativemessaging_manifests"]; ok {
		t.UpstreamNativeMessagingManifests = nil
		if paths, ok := value.([]string); ok {
			t.UpstreamNativeMessagingManifests = append([]string{}, paths...)
		} else if rawPaths, ok := value.([]any); ok {
			for _, rawPath := range rawPaths {
				if path, ok := rawPath.(string); ok && path != "" {
					t.UpstreamNativeMessagingManifests = append(t.UpstreamNativeMessagingManifests, path)
				}
			}
		}
		shouldInstallNativeHost = true
	}
	t.IncludeDefaultUpstreamNativeMessagingManifests = t.UpstreamNativeMessagingManifest == "" && len(t.UpstreamNativeMessagingManifests) == 0
	if nativeHostName, _ := config["upstream_nativemessaging_host_name"].(string); nativeHostName != "" {
		t.UpstreamNativeMessagingHostName = nativeHostName
		shouldInstallNativeHost = true
	}
	if waitTimeoutMS, ok := intFromConfig(config["upstream_nativemessaging_wait_timeout_ms"]); ok {
		t.WaitTimeoutMS = waitTimeoutMS
	}
	if extensionID, _ := config["injector_extension_id"].(string); extensionID != "" {
		t.ExtensionID = extensionID
		shouldInstallNativeHost = true
	}
	if userDataDir, _ := config["user_data_dir"].(string); userDataDir != "" && userDataDir != t.UserDataDir {
		t.setProfileUpstreamNativeMessagingManifests(userDataDir)
		t.UserDataDir = userDataDir
		shouldInstallNativeHost = true
	}
	if shouldInstallNativeHost && t.BoundPort != 0 {
		_ = t.installNativeHost(t.BoundPort)
	}
	if cdpURL, _ := config["cdp_url"].(string); cdpURL != "" {
		t.CDPURL = cdpURL
	}
}

func (t *NativeMessagingUpstreamTransport) GetServerConfig() map[string]any {
	if t.CDPURL == "" {
		return map[string]any{}
	}
	return map[string]any{"server_loopback_cdp_url": t.CDPURL}
}

func (t *NativeMessagingUpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{UpstreamNativeMessagingHostName: t.UpstreamNativeMessagingHostName}
}

func (t *NativeMessagingUpstreamTransport) Connect() error {
	t.stateMu.Lock()
	t.closed = false
	t.closeCh = make(chan struct{})
	t.stateMu.Unlock()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	t.Listener = listener
	t.BoundPort = listener.Addr().(*net.TCPAddr).Port
	t.URL = fmt.Sprintf("native://%s@127.0.0.1:%d", t.UpstreamNativeMessagingHostName, t.BoundPort)
	if err := t.installNativeHost(t.BoundPort); err != nil {
		_ = listener.Close()
		t.Listener = nil
		return err
	}
	go t.acceptLoop()
	return nil
}

func (t *NativeMessagingUpstreamTransport) Send(message map[string]any) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	conn := t.Conn
	if conn == nil {
		return fmt.Errorf("no native messaging peer is connected for %s", t.UpstreamNativeMessagingHostName)
	}
	return writeLengthPrefixedJSON(conn, message)
}

func (t *NativeMessagingUpstreamTransport) WaitForPeer() error {
	if t.Conn != nil {
		return nil
	}
	t.stateMu.Lock()
	closeCh := t.closeCh
	t.stateMu.Unlock()
	select {
	case <-t.peerCh:
		return nil
	case <-closeCh:
		return fmt.Errorf("native messaging transport for %s closed before a peer connected", t.UpstreamNativeMessagingHostName)
	case <-time.After(time.Duration(t.WaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("timed out waiting %dms for native messaging host %s", t.WaitTimeoutMS, t.UpstreamNativeMessagingHostName)
	}
}

func (t *NativeMessagingUpstreamTransport) Close() error {
	t.stateMu.Lock()
	t.closed = true
	closeCh := t.closeCh
	t.closeCh = make(chan struct{})
	t.stateMu.Unlock()
	close(closeCh)
	t.writeMu.Lock()
	if t.Conn != nil {
		_ = t.Conn.Close()
		t.Conn = nil
	}
	t.writeMu.Unlock()
	if t.Listener != nil {
		_ = t.Listener.Close()
		t.Listener = nil
	}
	t.peerCh = make(chan struct{})
	t.peerOnce = sync.Once{}
	return nil
}

func (t *NativeMessagingUpstreamTransport) Closed() bool {
	return t.closed
}

func (t *NativeMessagingUpstreamTransport) acceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			if !t.closed {
				t.emitClose(err)
			}
			return
		}
		go t.accept(conn)
	}
}

func (t *NativeMessagingUpstreamTransport) accept(conn net.Conn) {
	if t.Conn != nil && t.Conn != conn {
		_ = t.Conn.Close()
	}
	t.Conn = conn
	t.peerOnce.Do(func() { close(t.peerCh) })
	for {
		message, err := readLengthPrefixedJSON(conn)
		if err != nil {
			if t.Conn == conn {
				t.Conn = nil
				t.resetPeerWait()
			}
			if !t.closed {
				t.emitClose(err)
			}
			return
		}
		if messageType, _ := message["type"].(string); messageType == "modcdp.native.hello" {
			continue
		}
		t.emitRecv(message)
	}
}

func (t *NativeMessagingUpstreamTransport) resetPeerWait() {
	t.stateMu.Lock()
	defer t.stateMu.Unlock()
	t.peerCh = make(chan struct{})
	t.peerOnce = sync.Once{}
}

func (t *NativeMessagingUpstreamTransport) installNativeHost(port int) error {
	hostDir := filepath.Join(userHomeDir(), ".modcdp", "native-messaging")
	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		return err
	}
	configPath := filepath.Join(hostDir, t.UpstreamNativeMessagingHostName+".config.json")
	hostExecutablePath := filepath.Join(hostDir, t.UpstreamNativeMessagingHostName+".sh")
	if runtime.GOOS == "windows" {
		hostExecutablePath = filepath.Join(hostDir, t.UpstreamNativeMessagingHostName+".cmd")
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	configBody, err := json.MarshalIndent(map[string]any{"host": "127.0.0.1", "port": port}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(configPath, append(configBody, '\n'), 0o644); err != nil {
		return err
	}
	wrapper := fmt.Sprintf("#!/bin/sh\n%s=%s exec %s\n", nativeHostConfigEnv, shellQuote(configPath), shellQuote(exePath))
	if runtime.GOOS == "windows" {
		wrapper = fmt.Sprintf("@echo off\r\nset %s=%s\r\n%s\r\n", nativeHostConfigEnv, configPath, cmdQuote(exePath))
	}
	if err := os.WriteFile(hostExecutablePath, []byte(wrapper), 0o755); err != nil {
		return err
	}

	manifestPaths := []string{}
	if t.UpstreamNativeMessagingManifest != "" {
		manifestPaths = append(manifestPaths, t.UpstreamNativeMessagingManifest)
	}
	manifestPaths = append(manifestPaths, t.UpstreamNativeMessagingManifests...)
	if t.IncludeDefaultUpstreamNativeMessagingManifests || len(manifestPaths) == 0 {
		manifestPaths = append(manifestPaths, defaultUpstreamNativeMessagingManifestPaths(t.UpstreamNativeMessagingHostName)...)
	}
	manifestBody, err := json.MarshalIndent(map[string]any{
		"name":            t.UpstreamNativeMessagingHostName,
		"description":     "ModCDP Native Messaging bridge",
		"path":            hostExecutablePath,
		"type":            "stdio",
		"allowed_origins": []string{"chrome-extension://" + t.ExtensionID + "/"},
	}, "", "  ")
	if err != nil {
		return err
	}
	for _, manifestPath := range manifestPaths {
		if manifestPath == "" {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(manifestPath, append(manifestBody, '\n'), 0o644); err != nil {
			return err
		}
	}
	if runtime.GOOS == "windows" && len(manifestPaths) > 0 {
		if err := registerWindowsNativeMessagingHost(t.UpstreamNativeMessagingHostName, manifestPaths[0]); err != nil {
			return err
		}
	}
	return nil
}

func (t *NativeMessagingUpstreamTransport) setProfileUpstreamNativeMessagingManifests(userDataDir string) {
	previousProfileUpstreamNativeMessagingManifests := map[string]bool{}
	if t.UserDataDir != "" {
		previousProfileUpstreamNativeMessagingManifests[filepath.Join(t.UserDataDir, "NativeMessagingHosts", t.UpstreamNativeMessagingHostName+".json")] = true
		previousProfileUpstreamNativeMessagingManifests[filepath.Join(t.UserDataDir, "Default", "NativeMessagingHosts", t.UpstreamNativeMessagingHostName+".json")] = true
	}
	profileUpstreamNativeMessagingManifests := []string{
		filepath.Join(userDataDir, "NativeMessagingHosts", t.UpstreamNativeMessagingHostName+".json"),
		filepath.Join(userDataDir, "Default", "NativeMessagingHosts", t.UpstreamNativeMessagingHostName+".json"),
	}
	nextProfileUpstreamNativeMessagingManifests := map[string]bool{}
	for _, manifestPath := range profileUpstreamNativeMessagingManifests {
		nextProfileUpstreamNativeMessagingManifests[manifestPath] = true
	}
	filteredUpstreamNativeMessagingManifests := []string{}
	for _, manifestPath := range t.UpstreamNativeMessagingManifests {
		if previousProfileUpstreamNativeMessagingManifests[manifestPath] || nextProfileUpstreamNativeMessagingManifests[manifestPath] {
			continue
		}
		filteredUpstreamNativeMessagingManifests = append(filteredUpstreamNativeMessagingManifests, manifestPath)
	}
	t.UpstreamNativeMessagingManifests = append(profileUpstreamNativeMessagingManifests, filteredUpstreamNativeMessagingManifests...)
}

func defaultUpstreamNativeMessagingManifestPaths(nativeHostName string) []string {
	home := userHomeDir()
	if runtime.GOOS == "darwin" {
		return []string{
			filepath.Join(home, "Library/Application Support/Google/Chrome/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, "Library/Application Support/Google/Chrome Canary/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, "Library/Application Support/Google/ChromeForTesting/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, "Library/Application Support/Google/Chrome for Testing/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, "Library/Application Support/Google/Chrome SxS/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, "Library/Application Support/Chromium/NativeMessagingHosts", nativeHostName+".json"),
		}
	}
	if runtime.GOOS == "linux" {
		return []string{
			filepath.Join(home, ".config/google-chrome/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, ".config/google-chrome-for-testing/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, ".config/chromium/NativeMessagingHosts", nativeHostName+".json"),
			filepath.Join(home, ".config/chromium-browser/NativeMessagingHosts", nativeHostName+".json"),
		}
	}
	if runtime.GOOS == "windows" {
		return []string{filepath.Join(home, ".modcdp", "native-messaging", nativeHostName+".json")}
	}
	return nil
}

func registerWindowsNativeMessagingHost(nativeHostName string, manifestPath string) error {
	return osexec.Command(
		"reg",
		"add",
		`HKCU\Software\Google\Chrome\NativeMessagingHosts\`+nativeHostName,
		"/ve",
		"/t",
		"REG_SZ",
		"/d",
		manifestPath,
		"/f",
	).Run()
}

func runNativeMessagingHost(configPath string) error {
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	var config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return err
	}
	conn, err := net.Dial("tcp", net.JoinHostPort(config.Host, fmt.Sprint(config.Port)))
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := writeLengthPrefixedJSON(conn, map[string]any{"type": "modcdp.native.hello", "role": "native-host", "version": 1}); err != nil {
		return err
	}

	errCh := make(chan error, 2)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			message, err := readLengthPrefixedJSON(reader)
			if err != nil {
				errCh <- nil
				return
			}
			if err := writeLengthPrefixedJSON(conn, message); err != nil {
				errCh <- err
				return
			}
		}
	}()
	go func() {
		for {
			message, err := readLengthPrefixedJSON(conn)
			if err != nil {
				errCh <- nil
				return
			}
			if err := writeLengthPrefixedJSON(os.Stdout, message); err != nil {
				errCh <- err
				return
			}
		}
	}()
	return <-errCh
}

func writeLengthPrefixedJSON(writer io.Writer, message map[string]any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	var header [4]byte
	binary.LittleEndian.PutUint32(header[:], uint32(len(body)))
	if _, err := writer.Write(header[:]); err != nil {
		return err
	}
	_, err = writer.Write(body)
	return err
}

func readLengthPrefixedJSON(reader io.Reader) (map[string]any, error) {
	var header [4]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return nil, err
	}
	length := binary.LittleEndian.Uint32(header[:])
	body := make([]byte, length)
	if _, err := io.ReadFull(reader, body); err != nil {
		return nil, err
	}
	var message map[string]any
	if err := json.Unmarshal(body, &message); err != nil {
		return nil, err
	}
	return message, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func cmdQuote(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func userHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home
}
