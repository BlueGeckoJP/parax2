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
				if file.IsDir() || isImageFile(file.Name()) {
					children = append(children, filepath.Join(path, file.Name()))
				}
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
			} else {
				icon.SetResource(theme.MediaPhotoIcon())
				label.SetText(filepath.Base(id))
			}
		},
	)

	directoryTree.Root = "."
	directoryTree.OnSelected = func(id widget.TreeNodeID) {
	}

	directoryTreeLabel := widget.NewLabel("Directory Tree")

	leftPanel := container.New(layout.NewBorderLayout(directoryTreeLabel, nil, nil, nil), directoryTreeLabel, directoryTree)

	mainPanel := container.NewVBox()

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open Folder", func() {
				dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}
					if reader != nil {
						updateImageLists(mainPanel, reader.Path())
						directoryTree.Root = reader.Path()
					}
				}, myWindow)
			})),
	)

	myWindow.SetMainMenu(mainMenu)

	split := container.NewHSplit(leftPanel, mainPanel)
	split.SetOffset(0.2)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}

func updateImageLists(imageLists *fyne.Container, path string) {
	imageLists.RemoveAll()
	files, _ := os.ReadDir(path)

	addImage(files, path, imageLists, 0, 1)
}

func addImage(files []os.DirEntry, path string, imageLists *fyne.Container, depth int, maxDepth int) {
	list := container.NewHBox()

	for _, f := range files {
		if isImageFile(f.Name()) {
			image := canvas.NewImageFromFile(path + "/" + f.Name())
			image.SetMinSize(fyne.NewSize(200, 200))
			image.FillMode = canvas.ImageFillContain
			list.Add(image)
		} else if f.IsDir() && depth < maxDepth {
			subDir := filepath.Join(path, f.Name())
			subFiles, _ := os.ReadDir(subDir)
			addImage(subFiles, subDir, imageLists, depth+1, maxDepth)
		}
	}

	if list.Objects != nil {
		imageLists.Add(container.NewHScroll(list))
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
