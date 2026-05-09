package modcdp

import "testing"

func TestBorrowedExtensionInjectorReturnsNilWhenNoWorkerIsVisible(t *testing.T) {
	injector := NewBorrowedExtensionInjector(ExtensionInjectorConfig{
		Send: func(method string, params map[string]any, sessionID string) (map[string]any, error) {
			return map[string]any{"targetInfos": []any{}}, nil
		},
		ServiceWorkerReadyTimeoutMS: 1,
	})
	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("result = %#v", result)
	}
}
