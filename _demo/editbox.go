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

	erase()

	editbox := termwin.NewEditBox(25, 8, 16, 10, termwin.EditBoxWordWrap)

	editbox.InsertString("foobar01234567890")
	editbox.SetCursor(0, 0)
	editbox.InsertString("test\na\nb\nc\nd\n")
	editbox.SetCursor(-1, -1)
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

	editbox.SetCursor(0, 0)
	//editbox.Draw()

	for {
		termwin.Flush()
		termwin.Poll()
	}

	// termwin.Logln(editbox.Contents())

	// time.Sleep(time.Second * 1)
}

func erase() {
	buf := termbox.CellBuffer()
	for i := 0; i < len(buf); i++ {
		buf[i].Ch = '@'
	}
}
