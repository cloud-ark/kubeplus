package main

import (
        "fmt"
	"strings"
        postgresv1 "github.com/cloud-ark/kubeplus/postgres-crd-v2/pkg/apis/postgrescontroller/v1"
)

func getCreateUserCommands(desiredList []postgresv1.UserSpec) []string {
     var cmdList []string
     for _, user := range desiredList {
     	 username := user.User
	 password := user.Password 
     	 createUserCmd := strings.Fields("create user " + username + " with password '" + password + "';")
    	 var cmdString = strings.Join(createUserCmd, " ")
	 fmt.Printf("CreateUserCmd: %v\n", cmdString)
	 cmdList = append(cmdList, cmdString)
     }
     return cmdList
}

func getDropUserCommands(desiredList []postgresv1.UserSpec) []string {
     var cmdList []string
     for _, user := range desiredList {
     	 username := user.User
     	 dropUserCmd := strings.Fields("drop user " + username + ";")
    	 var cmdString = strings.Join(dropUserCmd, " ")
	 fmt.Printf("DropUserCmd: %v\n", cmdString)
	 cmdList = append(cmdList, cmdString)
     }
     return cmdList
}

func getAlterUserCommands(desiredList []postgresv1.UserSpec) []string {
     var cmdList []string
     for _, user := range desiredList {
     	 username := user.User
	 password := user.Password
     	 dropUserCmd := strings.Fields("alter user " + username +  " with password '" + password + "';")
    	 var cmdString = strings.Join(dropUserCmd, " ")
	 fmt.Printf("AlterUserCmd: %v\n", cmdString)
	 cmdList = append(cmdList, cmdString)
     }
     return cmdList
}

func getUserDiffList(desired []postgresv1.UserSpec, current []postgresv1.UserSpec) []postgresv1.UserSpec {
     var diffList []postgresv1.UserSpec
     for _, v := range desired {
     	 var found bool = false
     	 for _, v1 := range current {
	     if v.User == v1.User {
	     	found = true
	     }
	 }
	 if !found {
	    diffList = append(diffList, v)
	 }
     }
     //fmt.Printf("-- DiffList: %v--\n", diffList)
     return diffList
}

func getUserCommonList(desired []postgresv1.UserSpec, current []postgresv1.UserSpec) []postgresv1.UserSpec {
     var modifyList []postgresv1.UserSpec
     for _, v := range desired {
     	 for _, v1 := range current {
	     if v.User == v1.User {
	     	if v.Password != v1.Password {
		   modifyList = append(modifyList, v)
		}
	     }
	 }
     }
     //fmt.Printf("-- ModifyList: %v--\n", modifyList)
     return modifyList
}

func getUserCommands(desiredList []postgresv1.UserSpec, currentList []postgresv1.UserSpec) ([]string, []string, []string) {

     var createUserCommands []string
     var dropUserCommands []string
     var alterUserCommands []string

     if len(currentList) == 0 {
     	createUserCommands = getCreateUserCommands(desiredList)
     } else {
       	addList := getUserDiffList(desiredList, currentList)
	createUserCommands = getCreateUserCommands(addList)

       	dropList := getUserDiffList(currentList, desiredList)
	dropUserCommands = getDropUserCommands(dropList)

	alterList := getUserCommonList(desiredList, currentList)
	alterUserCommands = getAlterUserCommands(alterList)
     }
     return createUserCommands, dropUserCommands, alterUserCommands
}
