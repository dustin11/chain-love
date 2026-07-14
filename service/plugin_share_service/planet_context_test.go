package plugin_share_service

import (
	"testing"

	"senspace/domain/ds"
	"senspace/domain/sys"
)

func TestSelectShareSourcePlanetID(t *testing.T) {
	tests := []struct {
		name      string
		stored    int
		requested int
		want      int
		wantErr   bool
	}{
		{name: "uses stored relation", stored: 12, requested: 12, want: 12},
		{name: "rejects mismatched relation", stored: 12, requested: 13, wantErr: true},
		{name: "uses private scene id before relation exists", requested: 21, want: 21},
		{name: "rejects missing planet id", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := selectShareSourcePlanetID(test.stored, test.requested)
			if (err != nil) != test.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != test.want {
				t.Fatalf("planet id = %d, want %d", got, test.want)
			}
		})
	}
}

func TestIsCurrentSharePlanetOwner(t *testing.T) {
	share := &ds.PluginShare{CreatorUserId: 9, SourcePlanetId: 21}
	if !isCurrentSharePlanetOwner(sys.User{Id: 8, PlanetId: 21}, share) {
		t.Fatal("current stored planet owner should be allowed")
	}
	if isCurrentSharePlanetOwner(sys.User{Id: 9, PlanetId: 22}, share) {
		t.Fatal("creator must not override a mismatched stored planet relation")
	}
	if !isCurrentSharePlanetOwner(sys.User{Id: 9}, share) {
		t.Fatal("creator fallback should work before planet relations exist")
	}
	if isCurrentSharePlanetOwner(sys.User{Id: 8}, share) {
		t.Fatal("another user must not receive fallback owner permissions")
	}
}
