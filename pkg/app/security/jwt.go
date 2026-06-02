package security

import (
	"encoding/json"
	"log"
	"senspace/pkg/e"
	"senspace/pkg/setting"
	"senspace/pkg/setting/consts"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte(setting.Config.App.JwtSecret)

type JwtUser struct {
	Id          uint64 `json:"id"`
	Addr        string `json:"addr"`
	Nickname    string `json:"nickname,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	State       byte   `json:"state"`
	AccountPart byte   `json:"accountPart"`
	Country     string `json:"country" `
	City        string `json:"city"`
	PlanetId    int    `json:"planetId"`
	//应用
	Application string `json:"-" `
	//*jwt.StandardClaims
}

type jwtUserClaims struct {
	UserID      uint64 `json:"u"`
	Addr        string `json:"a,omitempty"`
	Nickname    string `json:"n,omitempty"`
	Avatar      string `json:"av,omitempty"`
	State       byte   `json:"s,omitempty"`
	AccountPart byte   `json:"ap,omitempty"`
	Country     string `json:"co,omitempty"`
	City        string `json:"ci,omitempty"`
	PlanetId    int    `json:"p,omitempty"`
	jwt.StandardClaims
}

func newJwtUserClaims(user JwtUser, now time.Time, expiresAt time.Time) jwtUserClaims {
	return jwtUserClaims{
		UserID:      user.Id,
		Addr:        user.Addr,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		State:       user.State,
		AccountPart: user.AccountPart,
		Country:     user.Country,
		City:        user.City,
		PlanetId:    user.PlanetId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt.Unix(),
			Id:        strconv.FormatUint(user.Id, 10),
			IssuedAt:  now.Unix(),
			Issuer:    user.Application,
			NotBefore: now.Unix(),
			Subject:   "login",
		},
	}
}

func (claims jwtUserClaims) toJwtUser() JwtUser {
	return JwtUser{
		Id:          claims.UserID,
		Addr:        claims.Addr,
		Nickname:    claims.Nickname,
		Avatar:      claims.Avatar,
		State:       claims.State,
		AccountPart: claims.AccountPart,
		Country:     claims.Country,
		City:        claims.City,
		PlanetId:    claims.PlanetId,
		Application: claims.Issuer,
	}
}

func GenerateToken(user JwtUser) (string, error) {
	nowTime := time.Now()
	// 使用统一配置的过期时间
	expiresTime := nowTime.Add(consts.AccessTokenTTL)
	tokenClaims := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		newJwtUserClaims(user, nowTime, expiresTime),
	)
	token, err := tokenClaims.SignedString(jwtSecret)

	return token, err
}

func ParseToken(token string) (JwtUser, error) {
	var user JwtUser
	tokenClaims, err := jwt.ParseWithClaims(token, &jwtUserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*jwtUserClaims); ok && tokenClaims.Valid {
			if claims.UserID > 0 || claims.Addr != "" {
				return claims.toJwtUser(), nil
			}
		}
	}

	if err != nil {
		return user, err
	}

	legacyClaims, legacyErr := parseLegacyToken(token)
	if legacyErr != nil {
		return user, legacyErr
	}
	if legacyClaims == nil {
		return user, nil
	}

	if unmarshalErr := json.Unmarshal([]byte(legacyClaims.Audience), &user); unmarshalErr != nil {
		e.PanicIfErr(unmarshalErr)
	}
	user.Application = legacyClaims.Issuer
	return user, nil
}

func parseLegacyToken(token string) (*jwt.StandardClaims, error) {
	legacyTokenClaims, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if legacyTokenClaims == nil {
		return nil, nil
	}
	claims, ok := legacyTokenClaims.Claims.(*jwt.StandardClaims)
	if !ok || !legacyTokenClaims.Valid {
		return nil, nil
	}
	if claims.Id != "" {
		if _, parseErr := strconv.ParseUint(claims.Id, 10, 64); parseErr != nil {
			log.Printf("ParseToken, fail to parse 'claims.Id': %v", parseErr)
		}
	}
	return claims, nil
}
