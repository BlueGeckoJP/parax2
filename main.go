package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

type Entry struct {
	Path     string
	Children []*Entry
	Depth    int
	isDir    bool
}

const (
	ViewModeList = iota
	ViewModeGrid
)

var imageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".svg":  true,
	".webp": true,
}
var backgroundRect *canvas.Rectangle
var thumbnailSize = fyne.NewSize(200, 200)
var maxDepth = 2
var wgMax = 10
var innerWGMax = 50

var entries []*Entry
var thumbnailCache = NewLRUCache(5000)
var currentPath = "."
var currentViewMode = ViewModeList
var config = loadConfig()

var directoryTree *widget.Tree
var directoryTreeLabel *widget.Label
var myWindow fyne.Window

func main() {
	myApp := app.New()
	myWindow = myApp.NewWindow("parax2")

	if config != nil {
		if config.ViewMode <= 1 {
			currentViewMode = config.ViewMode
		}
		if config.MaxDepth > 0 {
			maxDepth = config.MaxDepth
		}
		if config.CacheLimit > 0 {
			thumbnailCache = NewLRUCache(config.CacheLimit)
		}
		fmt.Println("Received raw config: ", config)
	}

	updateEntries(currentPath)

	backgroundRect = canvas.NewRectangle(color.Color(color.RGBA{51, 51, 51, 255}))
	backgroundRect.SetMinSize(thumbnailSize)

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
		var findId func([]*Entry)
		findId = func(entries []*Entry) {
			for _, entry := range entries {
				fmt.Println(entry.Path == id && !entry.isDir, entry)
				if entry.Path == id && !entry.isDir {
					openImageWithDefaultApp(entry.Path)
				}
				if entry.Children != nil {
					findId(entry.Children)
				}
			}
		}
		go findId(entries)
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
						updateMainPanel(mainPanel)
					}
				}, myWindow)
			})),
		fyne.NewMenu("View",
			fyne.NewMenuItem("List View", func() {
				currentViewMode = ViewModeList
				updateMainPanel(mainPanel)
			}),
			fyne.NewMenuItem("Grid View", func() {
				currentViewMode = ViewModeGrid
				updateMainPanel(mainPanel)
			}),
		),
	)

	myWindow.SetMainMenu(mainMenu)

	split := container.NewHSplit(leftPanel, container.NewVScroll(mainPanel))
	split.SetOffset(0.2)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}

func loadImageWithMmap(path string) (*ThumbnailWidget, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(stat.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	var img image.Image

	if filepath.Ext(path) == ".webp" {
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	} else {
		img, _, err = image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width > height {
		height = (height * int(thumbnailSize.Width)) / width
		width = int(thumbnailSize.Width)
	} else {
		width = (width * int(thumbnailSize.Height)) / height
		height = int(thumbnailSize.Height)
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	canvasImage := canvas.NewImageFromImage(dst)
	canvasImage.FillMode = canvas.ImageFillContain
	canvasImage.SetMinSize(thumbnailSize)

	thumbnail := newThumbnail(canvasImage, path)

	runtime.SetFinalizer(thumbnail, func(t *ThumbnailWidget) {
		syscall.Munmap(data)
	})

	return thumbnail, nil
}
func updateMainPanel(mainPanel *fyne.Container) {
	mainPanel.Objects = nil
	myWindow.SetTitle("Loading - parax2")
	wg := &WGWithCounter{
		wg:    sync.WaitGroup{},
		count: 0,
		max:   wgMax,
	}

	wg.Add(1, func() { addImage(entries, mainPanel, wg) })

	wg.wg.Wait()
	mainPanel.Refresh()
	myWindow.SetTitle("parax2")
	println("updateMainPanel done")
}

func addImage(entries []*Entry, mainPanel *fyne.Container, wg *WGWithCounter) {
	defer wg.Done()

	var c *fyne.Container
	switch currentViewMode {
	case ViewModeList:
		c = container.NewHBox()
	case ViewModeGrid:
		c = container.NewGridWrap(thumbnailSize)
	}

	innerWG := &WGWithCounter{
		wg:    sync.WaitGroup{},
		count: 0,
		max:   innerWGMax,
	}

	for _, entry := range entries {
		if entry.isDir {
			wg.Add(1, func() { addImage(entry.Children, mainPanel, wg) })
		} else {
			innerWG.Add(1, func() {
				go func() {
					defer innerWG.Done()
					image, exists := thumbnailCache.get(entry.Path)
					var thumbnail *ThumbnailWidget
					if !exists {
						var err error
						thumbnail, err = loadImageWithMmap(entry.Path)
						if err != nil {
							fmt.Println("Error loading image: ", err)
							return
						}
						thumbnailCache.add(entry.Path, thumbnail.Image)
					} else {
						thumbnail = newThumbnail(image, entry.Path)
					}
					c.Add(thumbnail)
				}()
			})
		}
	}

	go func() {
		innerWG.wg.Wait()

		if c.Objects != nil {
			relPath, _ := filepath.Rel(currentPath, filepath.Dir(entries[0].Path))
			go func() {
				if wg.count == 0 {
					mainPanel.Objects = append([]fyne.CanvasObject{container.NewVBox(
						widget.NewLabel(relPath),
						container.NewHScroll(c),
					)}, mainPanel.Objects...)
				} else {
					mainPanel.Objects = append([]fyne.CanvasObject{
						container.NewStack(
							backgroundRect,
							container.NewVBox(
								widget.NewLabel(relPath),
								container.NewHScroll(c),
							),
						),
					}, mainPanel.Objects...)
				}
			}()
		}
	}()
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return imageExts[ext]
}

func updateEntries(path string) {
	oldPath := currentPath
	currentPath = path
	if oldPath == path {
		return
	}
	entries = nil
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
	imageEntries := make([]*Entry, 0)

	for _, f := range files {
		if f.IsDir() && depth < maxDepth {
			p := filepath.Join(path, f.Name())
			children := addEntry(p, depth+1, maxDepth)
			if len(children) > 0 {
				entry := &Entry{
					Path:     p,
					Children: children,
					Depth:    depth,
					isDir:    true,
				}
				result = append(result, entry)
			}
		} else if isImageFile(f.Name()) {
			entry := &Entry{
				Path:     filepath.Join(path, f.Name()),
				Children: nil,
				Depth:    depth,
				isDir:    false,
			}
			imageEntries = append(imageEntries, entry)
		}
	}

	if len(imageEntries) > 0 {
		result = append(result, imageEntries...)
	}

	return result
}

func openImageWithDefaultApp(path string) {
	var cmd *exec.Cmd

	if config != nil && config.OpenCommand != nil {
		var placeholderIndex int
		for i, arg := range config.OpenCommand {
			if arg == "{}" {
				placeholderIndex = i
				break
			}
		}

		customCommand := make([]string, len(config.OpenCommand))
		copy(customCommand, config.OpenCommand)
		customCommand[placeholderIndex] = path
		cmd = exec.Command(customCommand[0], customCommand[1:]...)
	} else {
		cmd = exec.Command("xdg-open", path)
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command: ", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command: ", err)
		return
	}
}
