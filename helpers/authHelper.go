package helper

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var ErrUnauthorized = errors.New("unauthorized to access this resource")

func CheckUserType(ctx *gin.Context, role string) (err error) {
	userType := ctx.GetString("user_type")
	if userType != role {
		return ErrUnauthorized
	}
	return nil
}

func MatchUserTypeToUid(ctx *gin.Context, userId string) (err error) {
	userType := ctx.GetString("user_type")
	uid := ctx.GetString("uid")

	if userType == "USER" && uid != userId {
		return ErrUnauthorized
	}
	err = CheckUserType(ctx, userType)
	return err
}
