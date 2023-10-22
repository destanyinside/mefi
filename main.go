package main

import (
	"fmt"
	"github.com/destanyinside/mefi/cmd"
	"os"
)

var privateExitHandler = os.Exit

// ExitWrapper allow tests on main() exit values
func ExitWrapper(exit int) {
	privateExitHandler(exit)
}

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("%+v\n", err)
		ExitWrapper(1)
	}
}
