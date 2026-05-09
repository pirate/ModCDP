package modcdp

import "testing"

func TestDiscoveredExtensionInjectorReturnsNilWhenNoWorkerIsVisible(t *testing.T) {
	injector := NewDiscoveredExtensionInjector(ExtensionInjectorConfig{
		Send: func(method string, params map[string]any, sessionID string) (map[string]any, error) {
			return map[string]any{"targetInfos": []any{}}, nil
		},
	})
	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("result = %#v", result)
	}
}
