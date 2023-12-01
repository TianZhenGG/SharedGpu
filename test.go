package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/pratikms/fyne-syntax/go"
)

// createEditorWindow creates a new Fyne window with a text editor widget.
func createEditorWindow() fyne.Window {
	// Create a new Fyne application
	myApp := app.New()

	// Create a new window
	myWindow := myApp.NewWindow("Go Text Editor")

	// Create a new text editor widget
	editor := widget.NewMultiLineEntry()

	// Set the text editor's placeholder text
	editor.SetPlaceHolder("Start typing your Go code here...")

	// Enable syntax highlighting for Go keywords
	syntaxStyle := fyne_syntax.NewGoSyntaxStyle()
	editor.AddStyle(syntaxStyle)

	// Create a container to hold the text editor widget
	content := container.NewVBox(editor)

	// Set the content of the window to the container
	myWindow.SetContent(content)

	return myWindow
}

func main() {
	// Create a new Fyne window with a text editor
	window := createEditorWindow()

	// Show the window
	window.ShowAndRun()
}
