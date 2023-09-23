package usergroup

import (
	"github.com/sr-codefreak/user-group/db/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserGroup struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `bson:"name,omitempty"`
	MetaData map[string]any     `bson:"metaData,omitempty"`
	Users    []struct {
		ID    string `bson:"_id,omitempty"`
		Name  string `bson:"name,omitempty"`
		Email string `bson:"email,omitempty"`
		Phone string `bson:"phone,omitempty"`
	} `bson:"users,omitempty"`
	UserIds []string `bson:"userIds,omitempty"`
}

func (u UserGroupModel) CollectionName() string {
	return "userGroups"
}

type UserGroupModel struct {
	mongodb.UserGroup
	IdKey       string
	NameKey     string
	UsersKey    string
	UserIdsKey  string
	MetaDataKey string
}

var userGroupModel = &UserGroupModel{
	IdKey:       "_id",
	NameKey:     "name",
	UsersKey:    "users",
	UserIdsKey:  "userIds",
	MetaDataKey: "metaData",
}

func GetUserGroupModel() *UserGroupModel {
	return userGroupModel
}
