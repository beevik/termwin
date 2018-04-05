package main

import (
	"time"

	"github.com/nsf/termbox-go"

	"github.com/beevik/termwin"
)

func main() {
	termwin.CreateLog("out.log")

	err := termwin.Init()
	if err != nil {
		panic(err)
	}
	defer termwin.Close()

	erase()

	editbox := termwin.NewEditBox(5, 5, 16, 10, 0)

	editbox.InsertString("foobar01234567890")
	editbox.SetCursor(0, 0)
	editbox.InsertString("\ntest\na\nb\nc\nd")
	// editbox.SetCursor(0, 0)
	// editbox.InsertString("!!!")
	editbox.Draw()

	termwin.Flush()

	time.Sleep(time.Second * 1)
}

func erase() {
	buf := termbox.CellBuffer()
	for i := 0; i < len(buf); i++ {
		buf[i].Ch = '@'
	}
}
