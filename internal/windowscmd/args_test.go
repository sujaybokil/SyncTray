package windowscmd

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty", "", nil},
		{"whitespace", "  --no-upgrade\t--no-restart ", []string{"--no-upgrade", "--no-restart"}},
		{"quoted path", `--home "C:\Path With Spaces"`, []string{"--home", `C:\Path With Spaces`}},
		{"empty argument", `--label ""`, []string{"--label", ""}},
		{"escaped quote", `"say\"hello"`, []string{`say"hello`}},
		{"trailing backslash", `"C:\Path With Spaces\\"`, []string{`C:\Path With Spaces\`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SplitArgs(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SplitArgs(%q) = %#v, want %#v", tt.input, got, tt.want)
			}
		})
	}
}
