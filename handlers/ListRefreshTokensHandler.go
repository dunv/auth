package handlers

import (
	"fmt"
	"net/http"

	"github.com/dunv/uauth"
	"github.com/dunv/uhttp"
)

var ListRefreshTokensHandler = uhttp.NewHandler(
	uhttp.WithMiddlewares(uauth.AuthJWT()),
	uhttp.WithGet(func(r *http.Request, returnCode *int) interface{} {
		userService := uauth.NewUserService(uauth.UserDB(r), uauth.UserDBName(r))
		user, err := uauth.UserFromRequest(r)
		if err != nil {
			return err
		}

		tokens, err := userService.ListRefreshTokens(user.UserName, r.Context())
		if err != nil {
			return fmt.Errorf("could not find refreshTokens (%s)", err)
		}

		return map[string]interface{}{
			"refreshTokens": tokens,
		}
	}),
)
