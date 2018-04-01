package main

import "github.com/beevik/termwin"

func main() {
	err := termwin.Init()
	if err != nil {
		panic(err)
	}
	defer termwin.Close()

	editbox := termwin.NewEditBox(5, 5, 10, 3, 0)
	editbox.InsertRune(rune('a'))
}
