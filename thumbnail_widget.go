package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type ThumbnailWidget struct {
	widget.BaseWidget
	Image    *canvas.Image
	OnTapped func()
	Path     string
}

func (t *ThumbnailWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.Image)
}

func (t *ThumbnailWidget) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}

func newThumbnail(image *canvas.Image, path string) *ThumbnailWidget {
	t := &ThumbnailWidget{
		Image: image,
		OnTapped: func() {
			openImageWithDefaultApp(path)
		},
		Path: path,
	}
	t.ExtendBaseWidget(t)
	return t
}
