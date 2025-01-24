package main

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("parax2")

	directoryTree := widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			path := id
			if path == "" {
				path = "."
			}
			files, err := os.ReadDir(path)
			if err != nil {
				return nil
			}
			children := make([]widget.TreeNodeID, 0)
			for _, file := range files {
				children = append(children, filepath.Join(path, file.Name()))
			}
			return children
		},
		func(id widget.TreeNodeID) bool {
			if id == "" {
				return true
			}
			info, err := os.Stat(id)
			if err != nil {
				return false
			}
			return info.IsDir()
		},
		func(branch bool) fyne.CanvasObject {
			if branch {
				return container.NewHBox(
					widget.NewIcon(theme.FolderIcon()),
					widget.NewLabel(""),
				)
			}
			return container.NewHBox(
				widget.NewIcon(theme.FileImageIcon()),
				widget.NewLabel(""),
			)
		},
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			container := o.(*fyne.Container)
			icon := container.Objects[0].(*widget.Icon)
			label := container.Objects[1].(*widget.Label)

			if branch {
				icon.SetResource(theme.FolderIcon())
				label.SetText(filepath.Base(id))
			} else if isImageFile(id) {
				icon.SetResource(theme.FileImageIcon())
				label.SetText(filepath.Base(id))
			} else {
				icon.SetResource(theme.FileIcon())
				label.SetText(filepath.Base(id))
			}
		},
	)

	directoryTree.OnSelected = func(id widget.TreeNodeID) {
	}

	directoryTreeLabel := widget.NewLabel("Directory Tree")

	leftPanel := container.New(layout.NewBorderLayout(directoryTreeLabel, nil, nil, nil), directoryTreeLabel, directoryTree)

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

	split := container.NewHSplit(leftPanel, scroll)
	split.SetOffset(0.25)

	myWindow.SetContent(split)
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
