package uauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dunv/mongo"
	"github.com/dunv/uhttp"
)

type createUserModel struct {
	UserName  string   `bson:"userName" json:"userName"`
	FirstName string   `bson:"firstName,omitempty" json:"firstName,omitempty"`
	LastName  string   `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Password  string   `bson:"password" json:"password,omitempty"`
	Roles     []string `bson:"roles" json:"roles"`
}

var createUserHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Get User
	user := r.Context().Value(CtxKeyUser).(User)

	if !user.CheckPermission(CanCreateUsers) {
		uhttp.RenderError(w, r, fmt.Errorf("User does not have the required permission: %s", CanCreateUsers))
		return
	}

	// Parse requestedUserModel
	var userFromRequest createUserModel
	err := json.NewDecoder(r.Body).Decode(&userFromRequest)
	defer r.Body.Close()
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	// Get DB
	db := r.Context().Value(uhttp.CtxKeyDB).(*mongo.DbSession)

	// Verify all roles exist
	roleService := NewRoleService(db)
	allRoles, err := roleService.GetAllRoles()
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	verifiedRoles := []string{}
	for _, wantedRole := range userFromRequest.Roles {
		for _, existingRole := range *allRoles {
			if wantedRole == existingRole.Name {
				verifiedRoles = append(verifiedRoles, wantedRole)
			}
		}
	}

	if len(verifiedRoles) != len(userFromRequest.Roles) {
		uhttp.RenderError(w, r, fmt.Errorf("Not all desired roles for the new user are valid"))
		return
	}

	hashedPassword, _ := HashPassword(userFromRequest.Password)
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	userService := NewUserService(db)
	userToBeCreated := User{
		UserName:  userFromRequest.UserName,
		FirstName: userFromRequest.FirstName,
		LastName:  userFromRequest.LastName,
		Password:  &hashedPassword,
		Roles:     &verifiedRoles,
	}
	err = userService.CreateUser(&userToBeCreated)
	if err != nil {
		uhttp.RenderError(w, r, err)
		return
	}

	uhttp.RenderMessageWithStatusCode(w, r, 200, "Created successfully")
})

// CreateUserHandler <-
var CreateUserHandler = uhttp.Handler{
	Methods:      []string{"OPTIONS", "POST"},
	Handler:      createUserHandler,
	DbRequired:   true,
	AuthRequired: true,
}
