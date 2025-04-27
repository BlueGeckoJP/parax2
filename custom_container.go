package main

import "fyne.io/fyne/v2"

type CustomContainer interface {
	Add(*ThumbnailWidget)
	Title() string
	Objects() []fyne.CanvasObject
}
