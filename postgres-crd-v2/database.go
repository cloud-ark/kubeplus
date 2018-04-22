package main

import (
        "fmt"
	"strings"
)

func getDatabaseCommands(desiredList []string, currentList []string) ([]string, []string) {
     var createDatabaseCommands []string
     var deleteDatabaseCommands []string

     if len(currentList) == 0 {
     	createDatabaseCommands = getCreateDatabaseCommands(desiredList)
     } else {
	  addList := getDiffList(desiredList, currentList)
	  createDatabaseCommands = getCreateDatabaseCommands(addList)

	  dropList := getDiffList(currentList, desiredList)
	  deleteDatabaseCommands = getDropDatabaseCommands(dropList)
     }
     return createDatabaseCommands, deleteDatabaseCommands
}

func getCreateDatabaseCommands(dbList []string) []string {
     var cmdList []string
     for _, db := range dbList {
     	 createDBCmd := strings.Fields("create database " + db + ";")
    	 var cmdString = strings.Join(createDBCmd, " ")
	 fmt.Printf("CreateDBCmd: %v\n", cmdString)
	 cmdList = append(cmdList, cmdString)
     }
     return cmdList
}

func getDropDatabaseCommands(dbList []string) []string {
     var cmdList []string
     for _, db := range dbList {
     	 dropDBCmd := strings.Fields("drop database " + db + ";")
    	 var cmdString = strings.Join(dropDBCmd, " ")
	 fmt.Printf("DropDBCmd: %v\n", cmdString)
	 cmdList = append(cmdList, cmdString)
     }
     return cmdList
}

