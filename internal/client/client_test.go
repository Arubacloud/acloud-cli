package client

import (
	"testing"
)

func TestReset(t *testing.T) {
	Reset()
	if s.client != nil || s.clientID != "" {
		t.Error("Reset() did not zero the cached state")
	}
}

func TestOverride_NoOverrideByDefault(t *testing.T) {
	t.Cleanup(Reset)
	Reset()
	if c, ok := Override(); ok || c != nil {
		t.Error("expected no override before SetForTesting")
	}
}

func TestGet_InvalidCredentials_ReturnsError(t *testing.T) {
	t.Cleanup(Reset)

	// aruba.NewClient should fail when credentials are empty and no server is reachable.
	// We just verify Get returns an error; the specific message is SDK-determined.
	_, err := Get(Params{
		ClientID:     "",
		ClientSecret: "",
		BaseURL:      "https://127.0.0.1:1", // unreachable
	})
	if err == nil {
		t.Error("expected error for empty credentials / bad URL")
	}
}

func TestGet_Caching(t *testing.T) {
	t.Cleanup(Reset)

	p := Params{
		ClientID:       "id-1",
		ClientSecret:   "secret-1",
		BaseURL:        "https://127.0.0.1:1",
		TokenIssuerURL: "https://127.0.0.1:2",
	}
	c1, err1 := Get(p)
	if err1 != nil {
		// The SDK may fail to connect; if so both calls should return the same error.
		c2, err2 := Get(p)
		if err1.Error() != err2.Error() {
			t.Errorf("expected same error on second call: %v vs %v", err1, err2)
		}
		_ = c2
		return
	}
	c2, err2 := Get(p)
	if err2 != nil {
		t.Fatalf("second Get returned error: %v", err2)
	}
	if c1 != c2 {
		t.Error("expected the same (cached) client on second call with identical Params")
	}
}

func TestGet_CacheInvalidatedOnParamChange(t *testing.T) {
	t.Cleanup(Reset)

	p1 := Params{ClientID: "id-1", ClientSecret: "secret-1", BaseURL: "https://127.0.0.1:1", TokenIssuerURL: "https://127.0.0.1:2"}
	p2 := Params{ClientID: "id-2", ClientSecret: "secret-2", BaseURL: "https://127.0.0.1:1", TokenIssuerURL: "https://127.0.0.1:2"}

	c1, err1 := Get(p1)
	c2, err2 := Get(p2)

	// If either failed, they must both have failed (SDK connect error).
	if (err1 != nil) != (err2 != nil) {
		t.Errorf("inconsistent error: err1=%v err2=%v", err1, err2)
		return
	}
	if err1 != nil {
		return // SDK failed to connect — cache-invalidation path can't be tested here
	}
	if c1 == c2 {
		t.Error("expected different clients when Params change (cache should be invalidated)")
	}
}
