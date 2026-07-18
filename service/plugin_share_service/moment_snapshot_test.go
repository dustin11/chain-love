package plugin_share_service

import (
	"bytes"
	"os"
	"testing"

	"senspace/pkg/setting"
)

func TestMomentSnapshotCompressedRoundTrip(t *testing.T) {
	previous := setting.Config.App.RuntimeRootPath
	setting.Config.App.RuntimeRootPath = t.TempDir()
	t.Cleanup(func() { setting.Config.App.RuntimeRootPath = previous })

	want := bytes.Repeat([]byte(`{"schema":"senspace.planet-moment.v1"}`), 64)
	if err := writeMomentSnapshot("opaque-key", want); err != nil {
		t.Fatalf("writeMomentSnapshot: %v", err)
	}
	got, err := readMomentSnapshot("opaque-key")
	if err != nil {
		t.Fatalf("readMomentSnapshot: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatal("compressed snapshot round trip changed content")
	}
	if info, err := os.Stat(momentSnapshotPath("opaque-key")); err != nil || info.Size() >= int64(len(want)) {
		t.Fatalf("snapshot was not compressed: info=%v err=%v", info, err)
	}
}

func TestResolveExpiryDefaultsToPermanent(t *testing.T) {
	if got := resolveExpiry(0); got != nil {
		t.Fatalf("zero expiry = %v, want permanent", got)
	}
	if got := resolveExpiry(24); got == nil {
		t.Fatal("positive expiry must produce a deadline")
	}
}
