package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	flag1 := false
	flag2 := true
	
	if flag1 || flag2 {  // Line 117
		fmt.Println("Flag validation logic is incorrect.")
	}
	
	if !flag1 || flag2 {  // Line 135
		fmt.Println("Another flag validation logic is incorrect.")
	}
}