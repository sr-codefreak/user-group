package access

import (
	"github.com/sr-codefreak/user-group/db/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Access struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserId      string             `bson:"userId"`
	UserGroupId string             `bson:"userGroupId"`
	Roles       []string           `bson:"roles"`
}

type AccessModel struct {
	mongodb.UserGroup
	IdKey          string
	UserIdKey      string
	UserGroupIdKey string
	RolesKey       string
}

var accessModel = &AccessModel{
	IdKey:          "_id",
	UserIdKey:      "userId",
	UserGroupIdKey: "userGroupId",
	RolesKey:       "roles",
}

func GetModel() *AccessModel {
	return accessModel
}

func (u AccessModel) CollectionName() string {
	return "access"
}
