package ui

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	cases := []struct {
		name       string
		input      string
		defaultYes bool
		want       bool
		wantErr    error
	}{
		{name: "empty defaults to yes", input: "\n", defaultYes: true, want: true},
		{name: "empty defaults to no", input: "\n", defaultYes: false, want: false},
		{name: "y", input: "y\n", want: true},
		{name: "yes", input: "yes\n", want: true},
		{name: "uppercase YES", input: "YES\n", want: true},
		{name: "n", input: "n\n", defaultYes: true, want: false},
		{name: "no", input: "no\n", defaultYes: true, want: false},
		{name: "trims whitespace", input: "  y  \n", want: true},
		{name: "retries past garbage", input: "wat\ny\n", want: true},
		{name: "three invalid responses errors", input: "a\nb\nc\n"},
		{name: "stdin closed before answer", input: "", wantErr: ErrNoTTY},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			got, err := confirm(strings.NewReader(tc.input), out, "?", tc.defaultYes)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if tc.name == "three invalid responses errors" {
				if err == nil {
					t.Fatalf("expected error after %d invalid responses", maxAttempts)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConfirmHintReflectsDefault(t *testing.T) {
	cases := []struct {
		defaultYes bool
		wantHint   string
	}{
		{defaultYes: true, wantHint: "[Y/n]"},
		{defaultYes: false, wantHint: "[y/N]"},
	}
	for _, tc := range cases {
		out := &bytes.Buffer{}
		_, err := confirm(strings.NewReader("\n"), out, "proceed?", tc.defaultYes)
		if err != nil {
			t.Fatalf("defaultYes=%v: unexpected error: %v", tc.defaultYes, err)
		}
		if !strings.Contains(out.String(), tc.wantHint) {
			t.Errorf("defaultYes=%v: prompt %q missing hint %q", tc.defaultYes, out.String(), tc.wantHint)
		}
	}
}