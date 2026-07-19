package main

import "testing"

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
