package main

import (
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("parax2")

	imageHBox := container.NewHBox()

	files, _ := os.ReadDir("./")
	for _, f := range files {
		fmt.Println(f.Name())
		if !f.IsDir() && isImageFile(f.Name()) {
			fmt.Println("is image: ", f.Name())
			image := canvas.NewImageFromFile("./" + f.Name())
			image.Resize(fyne.NewSize(100, 100))
			image.FillMode = canvas.ImageFillOriginal
			imageHBox.Add(image)
		}
	}

	scroll := container.NewHScroll(imageHBox)

	myWindow.SetContent(scroll)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
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
