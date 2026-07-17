package auth_service

import (
	"senspace/domain/auth"
	"testing"
	"time"
)

func TestRefreshTokenGraceReuse(t *testing.T) {
	now := time.Now()
	base := auth.RefreshToken{
		Revoked:    true,
		ClientIp:   "127.0.0.1",
		UserAgent:  "browser",
		LastUsedAt: now.Add(-5 * time.Second),
	}

	tests := []struct {
		name      string
		record    *auth.RefreshToken
		clientIp  string
		userAgent string
		want      bool
	}{
		{name: "same client within grace", record: &base, clientIp: "127.0.0.1", userAgent: "browser", want: true},
		{name: "expired grace", record: withLastUsedAt(base, now.Add(-refreshTokenRotationGrace-time.Second)), clientIp: "127.0.0.1", userAgent: "browser"},
		{name: "different ip", record: &base, clientIp: "127.0.0.2", userAgent: "browser"},
		{name: "different user agent", record: &base, clientIp: "127.0.0.1", userAgent: "other"},
		{name: "active token", record: withRevoked(base, false), clientIp: "127.0.0.1", userAgent: "browser"},
		{name: "missing token", record: nil, clientIp: "127.0.0.1", userAgent: "browser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRefreshTokenGraceReuseAllowed(tt.record, tt.clientIp, tt.userAgent, now); got != tt.want {
				t.Fatalf("isRefreshTokenGraceReuseAllowed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func withLastUsedAt(record auth.RefreshToken, value time.Time) *auth.RefreshToken {
	record.LastUsedAt = value
	return &record
}

func withRevoked(record auth.RefreshToken, value bool) *auth.RefreshToken {
	record.Revoked = value
	return &record
}
