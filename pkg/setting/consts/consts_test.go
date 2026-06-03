package consts

import "testing"

func TestIsDevLikeEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{name: "dev", env: "dev", want: true},
		{name: "ldev", env: "ldev", want: true},
		{name: "wdev", env: "wdev", want: true},
		{name: "uat", env: "uat", want: false},
		{name: "prod", env: "prod", want: false},
		{name: "empty", env: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDevLikeEnv(tt.env); got != tt.want {
				t.Fatalf("IsDevLikeEnv(%q) = %v, want %v", tt.env, got, tt.want)
			}
		})
	}
}
