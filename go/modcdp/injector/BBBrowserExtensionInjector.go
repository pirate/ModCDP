package injector

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const DefaultBrowserbaseBaseURL = "https://api.browserbase.com"

type BBBrowserExtensionInjector struct {
	ExtensionInjector
	ExtensionID string
	ZipPath     string
	CleanupPath string
}

func NewBBBrowserExtensionInjector(options ExtensionInjectorConfig) BBBrowserExtensionInjector {
	return BBBrowserExtensionInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *BBBrowserExtensionInjector) Prepare() error {
	if i.Options.InjectorExtensionID != "" {
		i.ExtensionID = i.Options.InjectorExtensionID
		return nil
	}
	if i.ExtensionID != "" {
		return nil
	}
	extensionPath := i.Options.InjectorExtensionPath
	if extensionPath == "" {
		return nil
	} else if strings.HasSuffix(extensionPath, ".zip") {
		i.ZipPath = extensionPath
	} else {
		zipPath, cleanupPath, err := zipExtensionDir(extensionPath)
		if err != nil {
			return err
		}
		i.ZipPath = zipPath
		i.CleanupPath = cleanupPath
	}
	extensionID, err := i.uploadExtension(i.ZipPath)
	if err != nil {
		_ = i.Close()
		return err
	}
	i.ExtensionID = extensionID
	return nil
}

func (i *BBBrowserExtensionInjector) GetLauncherConfig() LaunchOptions {
	if i.ExtensionID == "" {
		return LaunchOptions{}
	}
	return LaunchOptions{InjectorExtensionID: i.ExtensionID}
}

func (i *BBBrowserExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	extensionID := i.Options.InjectorExtensionID
	i.Options.InjectorExtensionID = ""
	defer func() { i.Options.InjectorExtensionID = extensionID }()

	discovered, err := i.waitForReadyServiceWorker(i.Options.InjectorServiceWorkerReadyTimeoutMS, i.Options.InjectorTrustServiceWorkerTarget)
	if err != nil || discovered == nil {
		return discovered, err
	}
	discovered.Source = "bb"
	return discovered, nil
}

func (i *BBBrowserExtensionInjector) Close() error {
	if i.CleanupPath != "" {
		_ = os.RemoveAll(i.CleanupPath)
		i.CleanupPath = ""
	}
	return nil
}

func (i *BBBrowserExtensionInjector) uploadExtension(zipPath string) (string, error) {
	browserbaseAPIKey := firstNonEmptyString(i.Options.InjectorBrowserbaseAPIKey, os.Getenv("BROWSERBASE_API_KEY"))
	if browserbaseAPIKey == "" {
		return "", fmt.Errorf("BBBrowserExtensionInjector requires BROWSERBASE_API_KEY or launcher.launcher_options.browserbase_api_key")
	}
	baseURL := firstNonEmptyString(i.Options.InjectorBrowserbaseBaseURL, os.Getenv("BROWSERBASE_BASE_URL"), DefaultBrowserbaseBaseURL)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(zipPath))
	if err != nil {
		return "", err
	}
	file, err := os.Open(zipPath)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fileWriter, file); err != nil {
		_ = file.Close()
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	request, err := http.NewRequest("POST", strings.TrimRight(baseURL, "/")+"/v1/extensions", body)
	if err != nil {
		return "", err
	}
	request.Header.Set("X-BB-API-Key", browserbaseAPIKey)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(response.Body)
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return "", fmt.Errorf("Browserbase POST /v1/extensions -> %d%s", response.StatusCode, formatResponseBody(responseBody))
	}
	var payload map[string]any
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return "", err
	}
	extensionID, _ := payload["id"].(string)
	if extensionID == "" {
		return "", fmt.Errorf("Browserbase extension upload returned no id (got %s)", responseBody)
	}
	return extensionID, nil
}

func zipExtensionDir(extensionPath string) (string, string, error) {
	dir, err := os.MkdirTemp("", "modcdp-bb-extension-")
	if err != nil {
		return "", "", err
	}
	zipPath := filepath.Join(dir, "extension.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		_ = os.RemoveAll(dir)
		return "", "", err
	}
	archive := zip.NewWriter(zipFile)
	walkErr := filepath.WalkDir(extensionPath, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(extensionPath, path)
		if err != nil {
			return err
		}
		writer, err := archive.Create(relative)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
	closeArchiveErr := archive.Close()
	closeFileErr := zipFile.Close()
	if walkErr != nil || closeArchiveErr != nil || closeFileErr != nil {
		_ = os.RemoveAll(dir)
		if walkErr != nil {
			return "", "", walkErr
		}
		if closeArchiveErr != nil {
			return "", "", closeArchiveErr
		}
		return "", "", closeFileErr
	}
	return zipPath, dir, nil
}

func formatResponseBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	return ": " + string(body)
}
