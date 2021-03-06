package uauth

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"

	"github.com/dunv/uhttp"
	"github.com/dunv/ulog"
)

func AuthBasic(wantedUsername string, wantedMd5Password string) uhttp.Middleware {
	tmp := uhttp.Middleware(func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			user, pass, ok := r.BasicAuth()
			passMd5 := fmt.Sprintf("%x", md5.Sum([]byte(pass)))

			if !ok || user != wantedUsername || passMd5 != wantedMd5Password {
				packageConfig.UHTTP.RenderErrorWithStatusCode(w, r, http.StatusUnauthorized, fmt.Errorf("Unauthorized"), false)
				return
			}

			ctx := context.WithValue(r.Context(), CtxKeyUser, user)
			ctx = context.WithValue(ctx, CtxKeyAuthMethod, "basic")
			ulog.LogIfError(uhttp.AddLogOutput(w, "authMethod", "basic"))
			ulog.LogIfError(uhttp.AddLogOutput(w, "user", user))
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
	return tmp
}
