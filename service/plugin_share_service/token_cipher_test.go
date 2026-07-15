package plugin_share_service

import (
	"testing"

	"senspace/pkg/setting"
)

func TestManagementTokenRoundTrip(t *testing.T) {
	previous := setting.Config.App.JwtSecret
	setting.Config.App.JwtSecret = "test-share-secret"
	t.Cleanup(func() {
		setting.Config.App.JwtSecret = previous
	})

	ciphertext, err := encryptManagementToken("opaque-share-token")
	if err != nil {
		t.Fatal(err)
	}
	if ciphertext == "opaque-share-token" || ciphertext == "" {
		t.Fatalf("unexpected ciphertext: %q", ciphertext)
	}
	plaintext, err := decryptManagementToken(ciphertext)
	if err != nil {
		t.Fatal(err)
	}
	if plaintext != "opaque-share-token" {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestManagementTokenRequiresSecret(t *testing.T) {
	previous := setting.Config.App.JwtSecret
	setting.Config.App.JwtSecret = ""
	t.Cleanup(func() {
		setting.Config.App.JwtSecret = previous
	})

	if _, err := encryptManagementToken("opaque-share-token"); err == nil {
		t.Fatal("expected missing secret error")
	}
}
