package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"image/color"
)

type GridContainer struct {
	widget.BaseWidget
	Inner *fyne.Container
	Outer *fyne.Container
	title string
}

func newGridContainer(title string) *GridContainer {
	inner := container.NewGridWrap(fyne.NewSize(thumbnailWidth, thumbnailHeight))
	outer := container.NewStack(
		canvas.NewRectangle(color.RGBA{R: 51, G: 51, B: 51, A: 255}),
		container.NewVBox(
			widget.NewAccordion(
				widget.NewAccordionItem(
					title,
					inner,
				),
			),
		),
	)

	g := &GridContainer{
		Inner: inner, Outer: outer, title: title,
	}
	g.ExtendBaseWidget(g)

	return g
}

func (g *GridContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(g.Outer)
}

func (g *GridContainer) Add(t *ThumbnailWidget) {
	g.Inner.Add(t)
}

func (g *GridContainer) Title() string {
	return g.title
}

func (g *GridContainer) Objects() []fyne.CanvasObject {
	return g.Inner.Objects
}
