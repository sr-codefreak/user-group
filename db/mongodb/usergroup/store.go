package usergroup

import (
	"errors"

	"github.com/sr-codefreak/user-group/db/mongodb"
	"github.com/sr-codefreak/user-group/myerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserGroupStore interface {
	Create(group *UserGroup) error
	UpdateNameUpdateName(id primitive.ObjectID, name string) error
	AddUser(id primitive.ObjectID, userId primitive.ObjectID) error
	RemoveUser(id primitive.ObjectID, userId primitive.ObjectID) error
	DeleteById(ids primitive.ObjectID) error
	GetById(id string) (*UserGroup, error)
}

type userGroupStore struct{}

var UgStore = userGroupStore{}

func (userGroupStore) Create(group *UserGroup) error {
	_, err := mongodb.InsertOne(userGroupModel, group)
	if err != nil {
		return errors.Join(myerrors.ErrCreatingUserGroup, err)
	}
	return nil
}

func (userGroupStore) UpdateName(id primitive.ObjectID, name string) error {
	filter := bson.D{
		{Key: userGroupModel.IdKey, Value: id},
	}
	update := bson.D{
		bson.E{Key: userGroupModel.NameKey, Value: name},
	}
	_, err := mongodb.UpdateOne(userGroupModel, filter, update)
	if err != nil {
		return errors.Join(myerrors.ErrUpdatingUserGroupName, err)
	}
	return nil
}

func (userGroupStore) AddUser(id primitive.ObjectID, userId primitive.ObjectID) error {
	filter := bson.D{
		{Key: userGroupModel.IdKey, Value: id},
	}
	update := bson.D{
		// bson.E{Key: userGroupModel.NameKey, Value: name},
	}
	_, err := mongodb.UpdateOne(userGroupModel, filter, update)
	if err != nil {
		return errors.Join(myerrors.ErrUpdatingUserGroupName, err)
	}
	return nil
}

func (userGroupStore) GetById(id string) (*UserGroup, error) {
	ug := &UserGroup{}
	query := bson.D{
		bson.E{Key: userGroupModel.IdKey, Value: id},
	}
	exists, err := mongodb.FindOne(userGroupModel, ug, query)
	if !exists || err != nil {
		return nil, errors.Join(myerrors.ErrGetUserGroupById, err)
	}
	return ug, nil
}

func (userGroupStore) DeleteByIds(id primitive.ObjectID) error {

	query := bson.D{
		bson.E{Key: userGroupModel.IdKey, Value: id},
	}
	err := mongodb.DeleteOne(userGroupModel, query)
	if err != nil {
		return errors.Join(myerrors.ErrDeleteUserGroup, err)
	}
	return nil
}

func (userGroupStore) RemoveUser(id primitive.ObjectID, userId primitive.ObjectID) error {

	return nil
}
