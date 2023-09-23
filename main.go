package main

import (
	"fmt"

	"github.com/sr-codefreak/user-group/db/mongodb"
	"github.com/sr-codefreak/user-group/db/mongodb/usergroup"
)

func main() {

	// Connec to DB

	dbChan := make(chan struct{})
	mongodb.Connect("mongodb://localhost:27020", dbChan)
	<-dbChan

	// Create a user group

	ug := usergroup.UserGroup{
		Name: "My user group 2",
		MetaData: map[string]any{
			"desc":    "my group[ 2",
			"usecase": "abcdqqq",
		},
		UserIds: []string{
			"userid1",
			"userid2",
		},
	}

	err := usergroup.UgStore.Create(&ug)
	fmt.Print(err)

}
