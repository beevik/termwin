package main

import (
	"time"

	"github.com/nsf/termbox-go"

	"github.com/beevik/termwin"
)

func main() {
	err := termwin.Init()
	if err != nil {
		panic(err)
	}
	defer termwin.Close()

	buf := termbox.CellBuffer()
	for i := 0; i < len(buf); i++ {
		buf[i].Ch = '.'
	}

	editbox := termwin.NewEditBox(5, 5, 10, 8, 0)

	editbox.InsertString("foobar01234567890\ntest\na\nb\nc\nd")
	editbox.SetCursor(0, 0)
	editbox.InsertString("!!!")
	editbox.Draw()

	termwin.Flush()

	time.Sleep(time.Second * 3)
}
