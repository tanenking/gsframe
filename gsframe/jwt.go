package gsframe

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tanenking/gsframe/internal/logger"
)

type Claims struct {
	Data string `json:"data"`
	jwt.RegisteredClaims
}

const (
	DEFAULT_JWT_TIME_EXPIRE = time.Minute * 60 //token过期时间 1小时
)

func init() {
	// jwt.TimeFunc = GetNowTime
}

func EncodeJWT(data string, secret_key string, expire_time time.Duration) string {
	if len(data) <= 0 {
		return ""
	}
	now := time.Now() // GetNowTime()
	if expire_time <= 0 {
		expire_time = DEFAULT_JWT_TIME_EXPIRE
	}
	claims := &Claims{
		Data: data,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    GetProjectName(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expire_time)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed_token, err := token.SignedString([]byte(secret_key))
	if err != nil {
		logger.Log().Error("Error signing token: %+v", err)
		return ""
	}
	return signed_token
}

func DecodeJWT(jwt_string string, secret_key string) (valid bool, data string) {
	valid = false
	data = ""
	token, err := jwt.ParseWithClaims(jwt_string, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret_key), nil
	})
	if err != nil || !token.Valid {
		return
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || claims == nil {
		return
	}
	data = claims.Data
	if len(data) > 0 {
		valid = true
	}
	return
}
