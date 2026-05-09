package modcdp

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type LocalBrowserLaunchExtensionInjector struct {
	ExtensionInjector
	UnpackedExtensionPath string
	ExtensionID           string
	CleanupPath           string
}

func NewLocalBrowserLaunchExtensionInjector(options ExtensionInjectorConfig) LocalBrowserLaunchExtensionInjector {
	return LocalBrowserLaunchExtensionInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *LocalBrowserLaunchExtensionInjector) Prepare() error {
	extensionPath := i.Options.ExtensionPath
	if extensionPath == "" || i.UnpackedExtensionPath != "" {
		return nil
	}
	unpackedPath, cleanupPath, err := prepareUnpackedExtension(extensionPath, len(i.ExtensionRuntimeConfig()) > 0)
	if err != nil {
		return err
	}
	i.UnpackedExtensionPath = unpackedPath
	i.CleanupPath = cleanupPath
	if err := i.WriteExtensionRuntimeConfig(i.UnpackedExtensionPath); err != nil {
		return err
	}
	_, err = i.ResolveExtensionID()
	return err
}

func (i *LocalBrowserLaunchExtensionInjector) GetLauncherConfig() LaunchOptions {
	if i.UnpackedExtensionPath == "" {
		return LaunchOptions{}
	}
	return LaunchOptions{ExtraArgs: []string{"--load-extension=" + i.UnpackedExtensionPath}}
}

func (i *LocalBrowserLaunchExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	i.WakeConfiguredExtension()
	timeoutMS := i.Options.ServiceWorkerProbeTimeoutMS
	if timeoutMS > 3000 {
		timeoutMS = 3000
	}
	discovered, err := i.WaitForReadyServiceWorker(timeoutMS, i.Options.TrustMatchedServiceWorker)
	if err != nil || discovered == nil {
		return discovered, err
	}
	discovered.Source = "local_launch"
	return discovered, nil
}

func (i *LocalBrowserLaunchExtensionInjector) Close() error {
	if i.CleanupPath != "" {
		_ = os.RemoveAll(i.CleanupPath)
		i.CleanupPath = ""
	}
	return nil
}

func (i *LocalBrowserLaunchExtensionInjector) ResolveExtensionID() (string, error) {
	if i.ExtensionID != "" {
		return i.ExtensionID, nil
	}
	if i.Options.ExtensionID != "" {
		i.ExtensionID = i.Options.ExtensionID
	} else if i.UnpackedExtensionPath != "" {
		extensionID, err := extensionIDFromManifestKey(i.UnpackedExtensionPath)
		if err != nil {
			return "", err
		}
		i.ExtensionID = extensionID
	}
	if i.ExtensionID != "" {
		i.Options.ExtensionID = i.ExtensionID
	}
	return i.ExtensionID, nil
}

func prepareUnpackedExtension(extensionPath string, copyDirectory bool) (string, string, error) {
	if !strings.HasSuffix(extensionPath, ".zip") {
		if !copyDirectory {
			return extensionPath, "", nil
		}
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
	for _, file := range files {
		targetPath := filepath.Join(dir, file.Name)
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
