package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ListContainer struct {
	widget.BaseWidget
	Inner *fyne.Container
	Outer *fyne.Container
	title string
}

func newListContainer(title string) *ListContainer {
	inner := container.NewHBox()
	outer := container.NewStack(
		backgroundRect,
		container.NewVBox(
			widget.NewLabel(title),
			container.NewHScroll(inner),
		),
	)

	l := &ListContainer{
		Inner: inner, Outer: outer, title: title,
	}
	l.ExtendBaseWidget(l)

	return l
}

func (l *ListContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(l.Outer)
}

func (l *ListContainer) Add(t *ThumbnailWidget) {
	l.Inner.Add(t)
}

func (l *ListContainer) Title() string {
	return l.title
}

func (l *ListContainer) Objects() []fyne.CanvasObject {
	return l.Inner.Objects
}
