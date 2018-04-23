package main

import (
       "fmt"
       "strings"
)

func getCommandsToRun(actionHistory []string, setupCommands []string) []string {
     var commandsToRun []string
     for _, v := range setupCommands {
     	 var found bool = false
     	 for _, v1 := range actionHistory {
	     if v == v1 {
	     	found = true
	     }
	 }
	 if !found {
	    commandsToRun = append(commandsToRun, v)
	 }
     }
     fmt.Printf("-- commandsToRun: %v--\n", commandsToRun)     
     return commandsToRun
}

func getDiffList(desired []string, current []string) []string {
     var diffList []string
     for _, v := range desired {
     	 var found bool = false
     	 for _, v1 := range current {
	     if v == v1 {
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

func canonicalize(setupCommands1 []string) []string {
     var setupCommands []string
     //Convert setupCommands to Lower case
     for _, cmd := range setupCommands1 {
     	 setupCommands = append(setupCommands, strings.ToLower(cmd))
     }
  return setupCommands
}

func appendList(parentList *[]string, childList []string) {
     for _, val := range childList {
     	 *parentList = append(*parentList, val)
     }
}
