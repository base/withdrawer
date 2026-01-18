// Updated main.go file with fixed flag validation logic

package main

import (
	"fmt"
	"os"
)

func main() {
	customNetworkFlag := true
	standardNetworkFlag := false

	// Corrected logic: Change OR to AND for validation
	if customNetworkFlag && standardNetworkFlag {
		fmt.Println("Both custom and standard networks are enabled.")
	} else {
		fmt.Println("Validation logic corrected.")
	}
}
