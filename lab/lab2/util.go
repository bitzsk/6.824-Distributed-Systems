package raft

import "fmt"

// Debugging
const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		fmt.Printf(format, a...)
	}
	return
}

func (rf *Raft) DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		fmt.Printf("[ID:%s][T:%d][C:%d]",rf.serverId,rf.currentTerm,rf.count)
		fmt.Printf(format, a...)
		fmt.Println("")
	}
	return
}

func (rf *Raft) DPrintln(a ...interface{}) (n int, err error) {
	if Debug > 0 {
		fmt.Printf("[ID:%s][T:%d][C:%d]",rf.serverId,rf.currentTerm,rf.count)
		fmt.Println(a...)
	}
	return
}