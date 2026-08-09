package config

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		data string
		want Settings
	}{
		{"defaults", "# comment\nunknown=value", Settings{WebUIURL: DefaultWebUI}},
		{"trims values", " webui = http://localhost:9999 \n folder = C:\\Sync ", Settings{WebUIURL: "http://localhost:9999", SyncFolder: `C:\Sync`}},
		{"last valid web ui wins", "webui=\nwebui=http://localhost:9999", Settings{WebUIURL: "http://localhost:9999"}},
		{"empty syncthing clears override", "syncthing=C:\\old.exe\nsyncthing=", Settings{WebUIURL: DefaultWebUI}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.data); got != tt.want {
				t.Errorf("Parse() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestWithValue(t *testing.T) {
	tests := []struct{ data, key, value, want string }{
		{"", "syncthing", `C:\Tools\syncthing.exe`, "syncthing=C:\\Tools\\syncthing.exe\n"},
		{"webui=http://localhost:8384\n", "folder", `C:\Sync`, "webui=http://localhost:8384\nfolder=C:\\Sync\n"},
		{"webui=http://localhost:8384\nsyncthing=C:\\old.exe\n", "syncthing", `C:\Tools\syncthing.exe`, "webui=http://localhost:8384\nsyncthing=C:\\Tools\\syncthing.exe\n"},
	}
	for _, tt := range tests {
		if got := WithValue(tt.data, tt.key, tt.value); got != tt.want {
			t.Errorf("WithValue(%q, %q, %q) = %q, want %q", tt.data, tt.key, tt.value, got, tt.want)
		}
	}
}
