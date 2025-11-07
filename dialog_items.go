package main

// fileItem implements list.Item interface for the file dialog
type fileItem struct {
	name string
	path string
}

func (i fileItem) FilterValue() string { return i.name }
func (i fileItem) Title() string       { return i.name }
func (i fileItem) Description() string { return "" }

// bufferItem implements list.Item interface for the buffer dialog
type bufferItem struct {
	name  string
	index int
}

func (i bufferItem) FilterValue() string { return i.name }
func (i bufferItem) Title() string       { return i.name }
func (i bufferItem) Description() string { return "" }

// commandItem implements list.Item interface for the help dialog
type commandItem struct {
	command Command
}

func (i commandItem) FilterValue() string { return i.command.name + " " + i.command.description }
func (i commandItem) Title() string       { return i.command.name }
func (i commandItem) Description() string { return i.command.description }
