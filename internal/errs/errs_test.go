package errs

import (
	"errors"
	"strings"
	"testing"
)

func strPtr(s string) *string { return &s }

func TestFmtAPIError(t *testing.T) {
	cases := []struct {
		name    string
		status  int
		title   *string
		detail  *string
		want    string
		notWant string
	}{
		{
			name:    "nil title and detail",
			status:  404,
			want:    "API error (status 404)",
			notWant: ":",
		},
		{
			name:   "title only",
			status: 500,
			title:  strPtr("Bad Gateway"),
			want:   "API error (status 500): Bad Gateway",
		},
		{
			name:   "detail only",
			status: 502,
			detail: strPtr("upstream timeout"),
			want:   "API error (status 502) — upstream timeout",
		},
		{
			name:   "title and detail",
			status: 404,
			title:  strPtr("Not Found"),
			detail: strPtr("resource not found"),
			want:   "API error (status 404): Not Found — resource not found",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := FmtAPIError(tc.status, tc.title, tc.detail)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %q, want substring %q", err.Error(), tc.want)
			}
			if tc.notWant != "" && strings.Contains(err.Error(), tc.notWant) {
				t.Errorf("got %q, but should not contain %q", err.Error(), tc.notWant)
			}
		})
	}
}

func TestFromV2_NonHTTPError(t *testing.T) {
	plain := errors.New("plain error")
	got := FromV2(plain, false)
	if got != plain {
		t.Errorf("expected the original error to be returned unchanged, got %v", got)
	}
}

func TestFromV2_Nil(t *testing.T) {
	if got := FromV2(nil, false); got != nil {
		t.Errorf("FromV2(nil) = %v, want nil", got)
	}
}
