package user

import (
	"errors"

	"github.com/sr-codefreak/user-group/db/mongodb"
	"github.com/sr-codefreak/user-group/myerrors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStore interface {
	Create(u *User) error
	Update(u *User) error
	Delete(id primitive.ObjectID) error
	GetById(id string) (*User, error)
}

type userStore struct{}

func (userStore) Create(u *User) error {
	_, err := mongodb.InsertOne(userModel, u)
	if err != nil {
		return errors.Join(myerrors.ErrCreatingUser, err)
	}
	return nil
}

func (userStore) Update(u *User) error {
	filter := bson.D{
		{Key: userModel.IdKey, Value: u.ID},
	}
	update := bson.D{}
	if len(u.Name) > 0 {
		update = append(update, bson.E{Key: userModel.NameKey, Value: u.Name})
	}
	if len(u.Email) > 0 {
		update = append(update, bson.E{Key: userModel.EmailKey, Value: u.Email})
	}
	if len(u.Phone) > 0 {
		update = append(update, bson.E{Key: userModel.PhoneKey, Value: u.Phone})
	}
	_, err := mongodb.UpdateOne(userModel, filter, update)
	if err != nil {
		return errors.Join(myerrors.ErrUpdatingUser, err)
	}
	return nil
}

func (userStore) GetById(id string) (*User, error) {
	u := &User{}
	query := bson.D{
		bson.E{Key: userModel.IdKey, Value: id},
	}
	exists, err := mongodb.FindOne(userModel, u, query)
	if !exists || err != nil {
		return nil, errors.Join(myerrors.ErrGetUserGroupById, err)
	}
	return u, nil
}

func (userStore) Delete(id primitive.ObjectID) error {
	query := bson.D{
		bson.E{Key: userModel.IdKey, Value: id},
	}
	err := mongodb.DeleteOne(userModel, query)
	if err != nil {
		return errors.Join(myerrors.ErrDeleteUserGroup, err)
	}
	return nil
}
