package handlers

import (
	"fmt"
	"net/http"

	"github.com/dunv/uauth/config"
	"github.com/dunv/uauth/models"
	"github.com/dunv/uauth/permissions"
	"github.com/dunv/uauth/services"
	"github.com/dunv/uhttp"
	uhttpModels "github.com/dunv/uhttp/models"
	uhttpContextKeys "github.com/dunv/uhttp/contextkeys"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var deleteUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Get User
	user := r.Context().Value(config.CtxKeyUser).(models.User)

	if !user.CheckPermission(permissions.CanDeleteUsers) {
		uhttp.RenderError(w, r, fmt.Errorf("User does not have the required permission: %s", permissions.CanDeleteUsers))
		return
	}

	// Get Params
	params := r.Context().Value(uhttpContextKeys.CtxKeyParams).(map[string]interface{})

	// Get DB
	db := r.Context().Value(config.CtxKeyUserDB).(*mongo.Client)
	service := services.NewUserService(db)

	ID, err := primitive.ObjectIDFromHex(params["userId"].(string))
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	err = service.Delete(ID)
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	uhttp.RenderMessageWithStatusCode(w, r, 200, "Deleted successfully")
	return
})

var DeleteUserHandler = uhttpModels.Handler{
	DeleteHandler:             deleteUserHandler,
	AdditionalContextRequired: []uhttpModels.ContextKey{config.CtxKeyUserDB},
	AuthRequired:              true,
	RequiredParams: uhttpModels.Params{ParamMap: map[string]uhttpModels.ParamRequirement{
		"userId": uhttpModels.ParamRequirement{AllValues: true},
	}},
}
