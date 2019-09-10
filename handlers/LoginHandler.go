package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/dunv/uauth"
	"github.com/dunv/uauth/config"
	"github.com/dunv/uauth/models"
	"github.com/dunv/uauth/services"
	"github.com/dunv/uhttp"
	uhttpContextKeys "github.com/dunv/uhttp/contextkeys"
	uhttpModels "github.com/dunv/uhttp/models"
	"go.mongodb.org/mongo-driver/mongo"
)

type loginRequest struct {
	User models.User `json:"user"`
}

type loginResponse struct {
	User    models.User           `json:"user"`
	JWTUser models.UserWithClaims `json:"DO_NOT_USE_jwtUser"`
	JWT     string                `json:"jwt"`
}

var loginHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// Parse request
	loginRequest := loginRequest{}
	err := json.NewDecoder(r.Body).Decode(&loginRequest)
	defer r.Body.Close()
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	db := r.Context().Value(config.CtxKeyUserDB).(*mongo.Client)
	userService := services.NewUserService(db)
	userFromDb, err := userService.GetByUserName(loginRequest.User.UserName)

	// Verify user with password
	if err != nil || !(*userFromDb).CheckPassword(*loginRequest.User.Password) {
		uhttp.RenderError(w, r, fmt.Errorf("No user with this name/password exists (%s)", err))
		return
	}

	// Get Roles
	rolesService := services.NewRoleService(db)
	roles, err := rolesService.GetMultipleByName(*userFromDb.Roles)

	// Check error
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	permissions := models.MergeToPermissions(*roles)

	// Create jwt-token with the username set
	var userWithClaims = (*userFromDb).ToUserWithClaims()
	err = userWithClaims.UnmarshalAdditionalAttributes()
	if err != nil {
		uhttp.RenderError(w, r, fmt.Errorf("Could not unmarshal additonalAttributes (%s)", err))
		return
	}
	userWithClaims.IssuedAt = int64(time.Now().Unix())
	usedIssuer := uauth.Config().TokenIssuer
	userWithClaims.Issuer = usedIssuer
	userWithClaims.ExpiresAt = int64(time.Now().Unix() + 604800)
	userWithClaims.Permissions = &permissions
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, userWithClaims)

	bCryptSecret := r.Context().Value(uhttpContextKeys.CtxKeyBCryptSecret).(string)
	signedToken, err := token.SignedString([]byte(bCryptSecret))
	if err != nil {
		uhttp.RenderError(w, r, err)
	}

	// Add rolesDetails to user-model
	(*userFromDb).Permissions = &permissions

	// Clean
	(*userFromDb).Password = nil

	// Render response
	uhttp.Render(w, r, loginResponse{
		User:    *userFromDb,
		JWTUser: userWithClaims,
		JWT:     signedToken,
	})
})

// LoginHandler handler for getting JSON web token
var LoginHandler = uhttpModels.Handler{
	PostHandler:               loginHandler,
	AdditionalContextRequired: []uhttpModels.ContextKey{config.CtxKeyUserDB},
}