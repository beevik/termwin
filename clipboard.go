package termwin

var clipboard string

// ClipboardSet sets the contents of the clipboard.
func ClipboardSet(s string) {
	clipboard = s
}

// ClipboardClear clears the current contents of the clipboard.
func ClipboardClear() {
	clipboard = ""
}

// ClipboardGet returns the current contents of the clipboard
func ClipboardGet() string {
	return clipboard
}
