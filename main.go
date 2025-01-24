package main

import (
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("parax2")

	imageHBox := container.NewHBox()

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open Folder", func() {
				dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}
					if reader != nil {
						updateImageHBox(imageHBox, reader.Path())
					}
				}, myWindow)
			})),
	)

	myWindow.SetMainMenu(mainMenu)

	scroll := container.NewHScroll(imageHBox)

	myWindow.SetContent(scroll)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}

func updateImageHBox(imageHBox *fyne.Container, path string) {
	imageHBox.RemoveAll()
	files, _ := os.ReadDir(path)
	for _, f := range files {
		if !f.IsDir() && isImageFile(f.Name()) {
			image := canvas.NewImageFromFile(path + "/" + f.Name())
			image.Resize(fyne.NewSize(100, 100))
			image.FillMode = canvas.ImageFillOriginal
			imageHBox.Add(image)
		}
	}
}

func isImageFile(filename string) bool {
	extensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg"}
	for _, ext := range extensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}
