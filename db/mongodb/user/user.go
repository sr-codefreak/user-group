package user

import (
	"github.com/sr-codefreak/user-group/db/mongodb"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Email       string             `bson:"email"`
	Phone       string             `bson:"phone"`
	MetaData    map[string]any     `bson:"metaData"`
	UsersGroups []struct {
		ID   string `bson:"_id"`
		Name string `bson:"name"`
	} `bson:"usersGroups"`
	UserGroupIds []string `bson:"userGroupIds"`
}

type UserModel struct {
	mongodb.UserGroup
	IdKey         string
	NameKey       string
	EmailKey      string
	PhoneKey      string
	MetaDataKey   string
	UserGroupsKey string
	UsgidsKey     string
}

var userModel = &UserModel{
	IdKey:         "_id",
	NameKey:       "name",
	EmailKey:      "email",
	PhoneKey:      "phone",
	MetaDataKey:   "metaData",
	UserGroupsKey: "usersGroups",
	UsgidsKey:     "userGroupIds",
}

func GetUserGroupModel() *UserModel {
	return userModel
}

func (u UserModel) CollectionName() string {
	return "user"
}
