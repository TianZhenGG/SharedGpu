// HNZYjFVcG81ZW1vQ2hWS1pCOTRiaW5Cd0lIczF1c2F6TTkyMTRnNDBIcnlmNDFsSUFBQUFBJCQAAAAAAAAAAAEAAACm0-4~s~TLrrm10fjT4wAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAPLyZWXy8mVlO

package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Text Editor")

	label := widget.NewLabel("Hello, World!")

	w.SetContent(container.NewVBox(label))
	w.ShowAndRun()
}
