package myerrors

import "errors"

// ErrNoMongoConnection is returned when there is no mongo connection
var ErrNoMongoConnection = errors.New("mongo client is not connected")
var ErrCreatingUserGroup = errors.New("error creating user group")
var ErrUpdatingUserGroupName = errors.New("error updating user group name")
var ErrGetUserGroupById = errors.New("error getting user by Id")
var ErrDeleteUserGroup = errors.New("error deleting user group")

var (
	ErrCreatingUser = errors.New("error creating user")
	ErrUpdatingUser = errors.New("error updating user")
)
