package user_service

import (
	"chain-love/domain/sys"
	"chain-love/pkg/app"
	"chain-love/pkg/e"
)

func Login(addr, password string) (string, error) {
	u := (&sys.User{Addr: addr}).GetByAddr()
	e.PanicIf(u.Id == 0, "用户名或密码错误！") //帐号错误

	// if u.Password == util.MD5(password) {
	//生成token
	token, err := app.GenerateToken(u.ToJwtUser())
	return token, err
	// }
	// return "", errors.New("用户名或密码错误")
}
