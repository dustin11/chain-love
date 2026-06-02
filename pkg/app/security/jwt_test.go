package security

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestGenerateAndParseToken(t *testing.T) {
	source := JwtUser{
		Id:          42,
		Addr:        "0x1234567890abcdef1234567890abcdef12345678",
		Nickname:    "Tester",
		Avatar:      "avatar.png",
		State:       1,
		AccountPart: 3,
		Country:     "CN",
		City:        "Shanghai",
		PlanetId:    9,
		Application: "senspace",
	}

	token, err := GenerateToken(source)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	parsed, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if parsed.Id != source.Id ||
		parsed.Addr != source.Addr ||
		parsed.Nickname != source.Nickname ||
		parsed.Avatar != source.Avatar ||
		parsed.State != source.State ||
		parsed.AccountPart != source.AccountPart ||
		parsed.Country != source.Country ||
		parsed.City != source.City ||
		parsed.PlanetId != source.PlanetId ||
		parsed.Application != source.Application {
		t.Fatalf("parsed token mismatch: %#v", parsed)
	}
}

func TestParseLegacyAudienceToken(t *testing.T) {
	source := JwtUser{
		Id:          7,
		Addr:        "0xabcdefabcdefabcdefabcdefabcdefabcdefabcd",
		Nickname:    "Legacy",
		PlanetId:    3,
		Application: "senspace",
	}
	userData, err := json.Marshal(source)
	if err != nil {
		t.Fatalf("Marshal user failed: %v", err)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Audience:  string(userData),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		Id:        "7",
		IssuedAt:  time.Now().Unix(),
		Issuer:    source.Application,
		NotBefore: time.Now().Unix(),
		Subject:   "login",
	}).SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("legacy token sign failed: %v", err)
	}

	parsed, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken legacy failed: %v", err)
	}

	if parsed.Id != source.Id ||
		parsed.Addr != source.Addr ||
		parsed.Nickname != source.Nickname ||
		parsed.PlanetId != source.PlanetId ||
		parsed.Application != source.Application {
		t.Fatalf("parsed legacy token mismatch: %#v", parsed)
	}
}
