package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestParseWebUIConfig(t *testing.T) {
	tests := []struct {
		name string
		data string
		want string
	}{
		{
			name: "defaults when unset",
			data: "# personal settings\nother=value",
			want: defaultWebUI,
		},
		{
			name: "trims whitespace",
			data: "  webui=http://localhost:9999  ",
			want: "http://localhost:9999",
		},
		{
			name: "ignores empty values",
			data: "webui=\nname=SyncTray",
			want: defaultWebUI,
		},
		{
			name: "uses last non-empty value",
			data: "webui=http://127.0.0.1:8384\nwebui=http://localhost:9999",
			want: "http://localhost:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseWebUIConfig(tt.data); got != tt.want {
				t.Errorf("parseWebUIConfig() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseConfigReadsSyncFolder(t *testing.T) {
	settings := parseConfig("webui=http://localhost:9999\nfolder=C:\\Users\\Me\\Sync\nsyncthing=C:\\Tools\\syncthing.exe")
	if settings.webUIURL != "http://localhost:9999" {
		t.Fatalf("webUIURL = %q", settings.webUIURL)
	}
	if settings.syncFolder != "C:\\Users\\Me\\Sync" {
		t.Fatalf("syncFolder = %q", settings.syncFolder)
	}
	if settings.syncthingPath != "C:\\Tools\\syncthing.exe" {
		t.Fatalf("syncthingPath = %q", settings.syncthingPath)
	}
}

func TestParseConfigAllowsEmptySyncthingPath(t *testing.T) {
	settings := parseConfig("syncthing=\n")
	if settings.syncthingPath != "" {
		t.Errorf("syncthingPath = %q, want empty path for PATH discovery", settings.syncthingPath)
	}
}

func TestParseConfigIgnoresCommentsAndMalformedLines(t *testing.T) {
	settings := parseConfig("# webui=http://wrong\n; folder=C:\\wrong\nthis is not a setting\r\n webui = http://localhost:9999 \r\n")
	if settings.webUIURL != "http://localhost:9999" {
		t.Errorf("webUIURL = %q, want configured URL", settings.webUIURL)
	}
	if settings.syncFolder != "" {
		t.Errorf("syncFolder = %q, want empty", settings.syncFolder)
	}
}

func TestSyncthingArgsForVersion(t *testing.T) {
	if got, want := syncthingArgsForVersion("syncthing v1.29.7"), []string{"--no-browser", "--no-restart", "--no-upgrade"}; !reflect.DeepEqual(got, want) {
		t.Errorf("v1 args = %q, want %q", got, want)
	}
	if got, want := syncthingArgsForVersion("syncthing v2.1.1"), []string{"serve", "--no-browser", "--no-restart", "--no-upgrade"}; !reflect.DeepEqual(got, want) {
		t.Errorf("v2 args = %q, want %q", got, want)
	}
}

func TestSyncthingArgsForVersionHandlesUnexpectedOutput(t *testing.T) {
	got := syncthingArgsForVersion("Syncthing unknown build")
	want := []string{"serve", "--no-browser", "--no-restart", "--no-upgrade"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected-version args = %q, want %q", got, want)
	}
}

func TestMakeIconProducesICO(t *testing.T) {
	icon := makeIcon(statusRunning)
	if len(icon) < 6 {
		t.Fatal("icon is too short")
	}
	if icon[2] != 1 || icon[4] == 0 {
		t.Errorf("icon header = %v, want ICO type and one image", icon[:6])
	}
}

func TestStatusIconsAreDistinct(t *testing.T) {
	if reflect.DeepEqual(makeIcon(statusRunning), makeIcon(statusFailed)) {
		t.Error("running and failed icons must differ")
	}
	if reflect.DeepEqual(makeIcon(statusStarting), makeIcon(statusStopped)) {
		t.Error("starting and stopped icons must differ")
	}
}

func TestIsWebUIAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	if !isWebUIAvailable(server.URL) {
		t.Error("expected a responding web UI to be available")
	}
	server.Close()
	if isWebUIAvailable(server.URL) {
		t.Error("expected a closed test server to be unavailable")
	}
}

func TestWithConfigValue(t *testing.T) {
	got := withConfigValue("webui=http://localhost:8384\nsyncthing=C:\\Old\\syncthing.exe\n", "syncthing", "C:\\Tools\\syncthing.exe")
	want := "webui=http://localhost:8384\nsyncthing=C:\\Tools\\syncthing.exe\n"
	if got != want {
		t.Errorf("config = %q, want %q", got, want)
	}
}

func TestWithConfigValueAppendsAndHandlesEmptyConfig(t *testing.T) {
	if got, want := withConfigValue("", "syncthing", "C:\\Tools\\syncthing.exe"), "syncthing=C:\\Tools\\syncthing.exe\n"; got != want {
		t.Errorf("empty config = %q, want %q", got, want)
	}
	if got, want := withConfigValue("webui=http://localhost:8384\n", "folder", "C:\\Sync"), "webui=http://localhost:8384\nfolder=C:\\Sync\n"; got != want {
		t.Errorf("appended config = %q, want %q", got, want)
	}
}

func TestSplitWindowsArgs(t *testing.T) {
	got := splitWindowsArgs(`--home "C:\Users\Me\Scoop Config" --no-upgrade`)
	want := []string{"--home", `C:\Users\Me\Scoop Config`, "--no-upgrade"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitWindowsArgs() = %#v, want %#v", got, want)
	}
}

func TestSplitWindowsArgsHandlesEmptyAndEscapedQuotedArguments(t *testing.T) {
	got := splitWindowsArgs(`--label "" "C:\Path With Spaces\\" "say\"hello"`)
	want := []string{"--label", "", `C:\Path With Spaces\`, `say"hello`}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitWindowsArgs() = %#v, want %#v", got, want)
	}
}

func TestParseScoopShim(t *testing.T) {
	data := "path = \"C:\\Tools\\syncthing.exe\"\nargs = --home \"C:\\Tools\\config dir\" --no-upgrade\n"
	path, args, ok := parseScoopShim(data)
	if !ok {
		t.Fatal("expected Scoop metadata to parse")
	}
	if path != `C:\Tools\syncthing.exe` {
		t.Errorf("path = %q", path)
	}
	wantArgs := []string{"--home", `C:\Tools\config dir`, "--no-upgrade"}
	if !reflect.DeepEqual(args, wantArgs) {
		t.Errorf("args = %#v, want %#v", args, wantArgs)
	}
}
