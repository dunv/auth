package handlers

import (
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/dunv/uauth"
	"github.com/dunv/uhttp"
)

var GetUserHandler = uhttp.Handler{
	AddMiddleware: uauth.AuthJWT(),
	RequiredGet: uhttp.R{
		"userId": uhttp.STRING,
	},
	GetHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := uauth.UserFromRequest(r)
		if !user.CheckPermission(uauth.CanReadUsers) {
			uhttp.RenderError(w, r, fmt.Errorf("User does not have the required permission: %s", uauth.CanReadUsers))
			return
		}

		service := uauth.NewUserService(uauth.UserDB(r), uauth.UserDBName(r))
		ID, err := primitive.ObjectIDFromHex(*uhttp.GetAsString("userId", r))
		if err != nil {
			uhttp.RenderError(w, r, err)
			return
		}
		userFromDb, err := service.Get(ID)

		if err != nil {
			uhttp.RenderError(w, r, err)
			return
		}

		userFromDb.Password = nil
		uhttp.Render(w, r, *userFromDb)
	}),
}
