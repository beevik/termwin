package main

import (
	"fmt"

	"github.com/beevik/termwin"
)

func main() {
	fmt.Println("Testing...")

	err := termwin.Init()
	if err != nil {
		panic(err)
	}
	defer termwin.Close()

	editbox := termwin.NewEditBox(5, 5, 10, 3, 0)
	_ = editbox
}
