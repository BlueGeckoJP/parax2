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

type Entry struct {
	Path     string
	Children []*Entry
	Depth    int
	isDir    bool
}

const maxDepth = 2

var entries []*Entry
var currentPath = "."

var directoryTree *widget.Tree
var directoryTreeLabel *widget.Label

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("parax2")

	updateEntries(currentPath)

	directoryTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				children := make([]widget.TreeNodeID, 0)
				for _, entry := range entries {
					children = append(children, entry.Path)
				}
				return children
			}

			for _, entry := range entries {
				if entry.Path == id && entry.isDir {
					children := make([]widget.TreeNodeID, 0)
					for _, child := range entry.Children {
						children = append(children, child.Path)
					}
					return children
				}
			}
			return nil
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

	directoryTree.OnSelected = func(id widget.TreeNodeID) {
	}

	directoryTreeLabel = widget.NewLabel("Tree in " + currentPath)

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
						updateEntries(reader.Path())
						updateImageLists(mainPanel)
					}
				}, myWindow)
			})),
	)

	myWindow.SetMainMenu(mainMenu)

	split := container.NewHSplit(leftPanel, container.NewVScroll(mainPanel))
	split.SetOffset(0.2)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}

func updateImageLists(imageLists *fyne.Container) {
	imageLists.RemoveAll()

	addImage(entries, imageLists)
	reverseImageLists(imageLists)
}

func addImage(entries []*Entry, imageLists *fyne.Container) {
	list := container.NewHBox()

	for _, entry := range entries {
		if entry.isDir {
			addImage(entry.Children, imageLists)
		} else {
			image := canvas.NewImageFromFile(entry.Path)
			image.FillMode = canvas.ImageFillContain
			image.SetMinSize(fyne.NewSize(200, 200))
			list.Add(image)
		}
	}

	if list.Objects != nil {
		relPath, _ := filepath.Rel(currentPath, filepath.Dir(entries[0].Path))

		imageLists.Add(
			container.NewVBox(
				widget.NewLabel(relPath),
				container.NewHScroll(list),
			),
		)
	}
}

func reverseImageLists(imageLists *fyne.Container) {
	objects := imageLists.Objects
	for i := 0; i < len(objects)/2; i++ {
		j := len(objects) - i - 1
		objects[i], objects[j] = objects[j], objects[i]
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

func updateEntries(path string) {
	entries = nil
	currentPath = path
	result := addEntry(currentPath, 0, maxDepth)
	entries = result
	if directoryTreeLabel != nil {
		directoryTreeLabel.SetText("Tree in " + filepath.Base(currentPath))
	}
	if directoryTree != nil {
		directoryTree.Refresh()
	}
}

func addEntry(path string, depth int, maxDepth int) []*Entry {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	result := make([]*Entry, 0)

	for _, f := range files {
		if f.IsDir() && depth < maxDepth {
			p := filepath.Join(path, f.Name())
			entry := &Entry{
				Path:     p,
				Children: addEntry(p, depth+1, maxDepth),
				Depth:    depth,
				isDir:    true,
			}
			result = append(result, entry)
		} else if isImageFile(f.Name()) {
			entry := &Entry{
				Path:     filepath.Join(path, f.Name()),
				Children: nil,
				Depth:    depth,
				isDir:    false,
			}
			result = append(result, entry)
		}
	}
	return result
}
