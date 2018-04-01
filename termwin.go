package termwin

import termbox "github.com/nsf/termbox-go"

// Init must be called before any termwin controls can be used.
func Init() error {
	err := termbox.Init()
	if err != nil {
		return err
	}

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	return nil
}

// Close shuts down the termwin system.
func Close() {
	termbox.Close()
}
