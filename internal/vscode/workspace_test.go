package vscode

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	indicatorKey = "terminal.integrated.environmentChangesIndicator"
	relaunchKey  = "terminal.integrated.environmentChangesRelaunch"
)

func readDoc(t *testing.T, path string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read workspace: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("parse workspace: %v", err)
	}
	return doc
}

func sampleFolders() []Folder {
	return []Folder{{Name: "main", Path: ".worktrees/main"}}
}

func TestWriteWorkspace_NewFileSeedsTerminalDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.code-workspace")

	if err := WriteWorkspace(path, sampleFolders()); err != nil {
		t.Fatalf("WriteWorkspace: %v", err)
	}

	doc := readDoc(t, path)
	settings, ok := doc["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings missing or not an object: %#v", doc["settings"])
	}
	if got := settings[indicatorKey]; got != "off" {
		t.Errorf("%s = %v, want %q", indicatorKey, got, "off")
	}
	if got := settings[relaunchKey]; got != false {
		t.Errorf("%s = %v, want false", relaunchKey, got)
	}
}

func TestWriteWorkspace_ExistingFilePreservesUserSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.code-workspace")
	initial := map[string]any{
		"folders": []Folder{{Name: "old", Path: ".worktrees/old"}},
		"settings": map[string]any{
			indicatorKey:        "on",
			relaunchKey:         true,
			"editor.fontSize":   14,
			"workbench.colorTheme": "Default Dark+",
		},
	}
	raw, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteWorkspace(path, sampleFolders()); err != nil {
		t.Fatalf("WriteWorkspace: %v", err)
	}

	doc := readDoc(t, path)
	settings, ok := doc["settings"].(map[string]any)
	if !ok {
		t.Fatalf("settings missing: %#v", doc["settings"])
	}
	if got := settings[indicatorKey]; got != "on" {
		t.Errorf("%s overwritten: got %v, want %q", indicatorKey, got, "on")
	}
	if got := settings[relaunchKey]; got != true {
		t.Errorf("%s overwritten: got %v, want true", relaunchKey, got)
	}
	if got := settings["editor.fontSize"]; got != float64(14) {
		t.Errorf("editor.fontSize = %v, want 14", got)
	}
	if got := settings["workbench.colorTheme"]; got != "Default Dark+" {
		t.Errorf("colorTheme dropped: got %v", got)
	}
}

func TestWriteWorkspace_ExistingFileDoesNotReinjectDeletedKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.code-workspace")
	initial := map[string]any{
		"folders":  []Folder{{Name: "old", Path: ".worktrees/old"}},
		"settings": map[string]any{"editor.fontSize": 14},
	}
	raw, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteWorkspace(path, sampleFolders()); err != nil {
		t.Fatalf("WriteWorkspace: %v", err)
	}

	settings := readDoc(t, path)["settings"].(map[string]any)
	if _, present := settings[indicatorKey]; present {
		t.Errorf("%s was re-injected after user removed it", indicatorKey)
	}
	if _, present := settings[relaunchKey]; present {
		t.Errorf("%s was re-injected after user removed it", relaunchKey)
	}
	if got := settings["editor.fontSize"]; got != float64(14) {
		t.Errorf("editor.fontSize lost: got %v", got)
	}
}

func TestWriteWorkspace_PreservesUnknownTopLevelKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.code-workspace")
	initial := map[string]any{
		"folders":    []Folder{{Name: "old", Path: ".worktrees/old"}},
		"extensions": map[string]any{"recommendations": []any{"golang.go"}},
		"tasks":      map[string]any{"version": "2.0.0"},
	}
	raw, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	newFolders := []Folder{{Name: "feat", Path: ".worktrees/feat"}}
	if err := WriteWorkspace(path, newFolders); err != nil {
		t.Fatalf("WriteWorkspace: %v", err)
	}

	doc := readDoc(t, path)
	gotFolders, _ := json.Marshal(doc["folders"])
	wantFolders, _ := json.Marshal(newFolders)
	if !reflect.DeepEqual(gotFolders, wantFolders) {
		t.Errorf("folders not updated: got %s, want %s", gotFolders, wantFolders)
	}
	if _, ok := doc["extensions"]; !ok {
		t.Errorf("extensions key dropped")
	}
	if _, ok := doc["tasks"]; !ok {
		t.Errorf("tasks key dropped")
	}
}

func TestWriteWorkspace_ExistingFileWithoutSettingsStaysWithoutSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.code-workspace")
	initial := map[string]any{
		"folders": []Folder{{Name: "old", Path: ".worktrees/old"}},
	}
	raw, _ := json.MarshalIndent(initial, "", "  ")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if err := WriteWorkspace(path, sampleFolders()); err != nil {
		t.Fatalf("WriteWorkspace: %v", err)
	}

	doc := readDoc(t, path)
	if _, present := doc["settings"]; present {
		t.Errorf("settings was injected on existing file without settings: %#v", doc["settings"])
	}
}
