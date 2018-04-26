package main

import (
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

	editbox := termwin.NewEditBox(25, 8, 16, 10, termwin.EditBoxWordWrap)

	editbox.InsertString("foobar01234567890")
	editbox.CursorSet(0, 0)
	editbox.InsertString("00000\n11111\n22222\n33333\n4\n5\n")
	editbox.CursorSet(-1, -1)
	editbox.InsertString("lalala\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("fofofo\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("babababa\n")
	editbox.InsertString("end")

	editbox.CursorSet(0, 0)

	for {
		termwin.Flush()
		termwin.Poll()
	}
}

func erase() {
	buf := termbox.CellBuffer()
	for i := 0; i < len(buf); i++ {
		buf[i].Ch = '@'
	}
}
