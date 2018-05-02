package main

import (
	"fmt"
	"strings"

	"github.com/beevik/termwin"
)

func main() {
	termwin.CreateLog("out.log")

	err := termwin.Init()
	if err != nil {
		panic(err)
	}

	x, y := termwin.Size()

	editbox := termwin.NewEditBox(1, 1, x-2, y-2, 0)

	editbox.InsertString(strings.Repeat("-", x-3) + "\n")
	for i := 1; i <= 50; i++ {
		editbox.InsertString(fmt.Sprintf("Line %d\n", i))
	}
	editbox.InsertString(strings.Repeat("-", x-3) + "\n")

	editbox.CursorSet(0, 0)

	for {
		termwin.Flush()
		if termwin.Poll() != nil {
			break
		}
	}

	termwin.Close()

	fmt.Println(editbox.Contents())
}
