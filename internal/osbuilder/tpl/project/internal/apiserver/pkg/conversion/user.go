package conversion

import (
	"github.com/onexstack/onexstack/pkg/core"

    "{{.M.ModuleName}}/internal/{{.Web.Name}}/model"
	{{.Web.APIImportPath}}
)

// UserMToUserV1 converts a UserM object to a User object in the v1 API format.
func UserMToUserV1(userM *model.UserM) *{{.M.APIAlias}}.User {
	var user {{.M.APIAlias}}.User
    _ = core.CopyWithConverters(&user, userM)
	return &user
}

// UserV1ToUserM converts a User object from the v1 API format to UserM object.
func UserV1ToUserM(user *{{.M.APIAlias}}.User) *model.UserM {
	var userM model.UserM
	_ = core.CopyWithConverters(&userM, user)
	return &userM
}
