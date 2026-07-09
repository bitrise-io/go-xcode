package xcscheme

import (
	"errors"
	"testing"
)

func TestSchemeNotFoundError_Error(t *testing.T) {
	err := NotFoundError{Scheme: "Scheme", Container: "Workspace"}
	want := "scheme Scheme not found in Workspace"
	if err.Error() != want {
		t.Errorf("SchemeNotFoundError.Error() = %v, want %v", err, want)
	}
}

func TestIsSchemeNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "SchemeNotFoundError",
			err:  NotFoundError{Scheme: "Scheme", Container: "Workspace"},
			want: true,
		},
		{
			name: "not SchemeNotFoundError",
			err:  errors.New("other error"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsSchemeNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}
