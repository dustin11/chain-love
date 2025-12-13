package security

import (
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/pkg/setting/consts"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte(setting.Config.App.JwtSecret)

type JwtUser struct { // todo 增加字段
	Id uint64 `json:"id"`
	// Addr     string
	Nickname    string `json:"nickname,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	State       byte   `json:"state"`
	AccountPart byte   `json:"accountPart"`
	// domain.AreaModel
	Country string `json:"country" `
	// Province string
	City string `json:"city"`
	//应用
	Application string `json:"-" `
	//*jwt.StandardClaims
}

func GenerateToken(user JwtUser) (string, error) {
	nowTime := time.Now()
	// 使用统一配置的过期时间
	expiresTime := nowTime.Add(consts.AccessTokenTTL)
	userData, err := json.Marshal(user)
	e.PanicIfErr(err)
	claims := jwt.StandardClaims{
		Audience:  string(userData),                // 受众
		ExpiresAt: expiresTime.Unix(),              // 失效时间
		Id:        strconv.FormatUint(user.Id, 10), // 编号
		IssuedAt:  nowTime.Unix(),                  // 签发时间
		Issuer:    user.Application,                // 签发人
		NotBefore: nowTime.Unix(),                  // 生效时间
		Subject:   "login",                         // 主题
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)

	return token, err
}

func ParseToken(token string) (JwtUser, error) {
	var user JwtUser
	tokenClaims, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*jwt.StandardClaims); ok && tokenClaims.Valid {
			//id, err := strconv.ParseUint(claims.Id, 10, 64)
			if err != nil {
				log.Printf("ParseToken, fail to parse 'claims.Id': %v", err)
			} else {
				er := json.Unmarshal([]byte(claims.Audience), &user)
				e.PanicIfErr(er)
				//user = JwtUser{Id: id, Username: claims.Audience, Application: claims.Issuer}
			}
		}
	}

	return user, err
}
