package utility

import "testing"

func Test_globEscapePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Escaping test",
			args: args{path: "[]-?*"},
			want: "\\[\\]\\-\\?\\*",
		},
		// filepath.Glob(1) does not match for `test\.xcodeproj` if go version 1.8.3
		{
			name: "'.' is not escaped",
			args: args{path: "test.xcodeproj"},
			want: "test.xcodeproj",
		},
		{
			name: "`\\` in path",
			args: args{path: "\\"},
			want: "\\\\",
		},
		{
			name: "`\\` with",
			args: args{path: "\\[?"},
			want: "\\\\\\[\\?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeGlobPath(tt.args.path); got != tt.want {
				t.Errorf("globEscapePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
