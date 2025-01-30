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
)

type Entry struct {
	Path     string
	Children []*Entry
	Depth    int
	isDir    bool
}

type CacheNode struct {
	key   string
	image *canvas.Image
	prev  *CacheNode
	next  *CacheNode
}

type LRUCache struct {
	capacity int
	cache    map[string]*CacheNode
	head     *CacheNode
	tail     *CacheNode
}

type ThumbnailWidget struct {
	widget.BaseWidget
	Image    *canvas.Image
	OnTapped func()
}

const maxDepth = 2
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
}
var backgroundRect *canvas.Rectangle
var thumbnailSize = fyne.NewSize(200, 200)

var entries []*Entry
var thumbnailCache = NewLRUCache(5000)
var currentPath = "."
var currentViewMode = ViewModeList

var directoryTree *widget.Tree
var directoryTreeLabel *widget.Label

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("parax2")

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

func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*CacheNode),
	}
}

func (c *LRUCache) moveToFront(node *CacheNode) {
	if node == c.head {
		return
	}
	if node == c.tail {
		c.tail = node.prev
		c.tail.next = nil
	} else if node.prev != nil {
		node.prev.next = node.next
		node.next.prev = node.prev
	}
	node.prev = nil
	node.next = c.head
	if c.head != nil {
		c.head.prev = node
	}
	c.head = node
}

func (c *LRUCache) add(key string, image *canvas.Image) {
	if node, exists := c.cache[key]; exists {
		node.image = image
		c.moveToFront(node)
		return
	}

	node := &CacheNode{key, image, nil, nil}
	c.cache[key] = node

	if c.head == nil {
		c.head = node
		c.tail = node
	} else {
		node.next = c.head
		c.head.prev = node
		c.head = node
	}

	if len(c.cache) > c.capacity {
		delete(c.cache, c.tail.key)
		c.tail = c.tail.prev
		if c.tail != nil {
			c.tail.next = nil
		}
	}
}

func (c *LRUCache) get(key string) (*canvas.Image, bool) {
	if node, exists := c.cache[key]; exists {
		c.moveToFront(node)
		return node.image, true
	}
	return nil, false
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

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
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
	switch currentViewMode {
	case ViewModeList:
		addImageHBox(entries, mainPanel)
	case ViewModeGrid:
		addImageGrid(entries, mainPanel)
	}
}

func addImageHBox(entries []*Entry, mainPanel *fyne.Container) {
	list := container.NewHBox()

	for _, entry := range entries {
		if entry.isDir {
			addImageHBox(entry.Children, mainPanel)
		} else {
			image, exists := thumbnailCache.get(entry.Path)
			var thumbnail *ThumbnailWidget
			if !exists {
				var err error
				thumbnail, err = loadImageWithMmap(entry.Path)
				if err != nil {
					fmt.Println("Error loading image: ", err)
					continue
				}
				thumbnailCache.add(entry.Path, image)
			} else {
				thumbnail = newThumbnail(image, entry.Path)
			}
			list.Add(thumbnail)
		}
	}

	if list.Objects != nil {
		relPath, _ := filepath.Rel(currentPath, filepath.Dir(entries[0].Path))
		go func() {
			objLen := len(mainPanel.Objects)

			if objLen%2 == 0 {
				mainPanel.Objects = append([]fyne.CanvasObject{container.NewVBox(
					widget.NewLabel(relPath),
					container.NewHScroll(list),
				)}, mainPanel.Objects...)
			} else {
				mainPanel.Objects = append([]fyne.CanvasObject{
					container.NewStack(
						backgroundRect,
						container.NewVBox(
							widget.NewLabel(relPath),
							container.NewHScroll(list),
						),
					),
				}, mainPanel.Objects...)
			}
		}()
	}
}

func addImageGrid(entries []*Entry, mainPanel *fyne.Container) {
	grid := container.NewGridWrap(thumbnailSize)

	for _, entry := range entries {
		if entry.isDir {
			addImageGrid(entry.Children, mainPanel)
		} else {
			image, exists := thumbnailCache.get(entry.Path)
			var thumbnail *ThumbnailWidget
			if !exists {
				var err error
				thumbnail, err = loadImageWithMmap(entry.Path)
				if err != nil {
					fmt.Println("Error loading image: ", err)
					continue
				}
				thumbnailCache.add(entry.Path, image)
			} else {
				thumbnail = newThumbnail(image, entry.Path)
			}
			grid.Add(thumbnail)
		}
	}

	if grid.Objects != nil {
		relPath, _ := filepath.Rel(currentPath, filepath.Dir(entries[0].Path))
		go func() {
			objLen := len(mainPanel.Objects)

			if objLen%2 == 0 {
				mainPanel.Objects = append([]fyne.CanvasObject{container.NewVBox(
					widget.NewLabel(relPath),
					grid,
				)}, mainPanel.Objects...)
			} else {
				mainPanel.Objects = append([]fyne.CanvasObject{
					container.NewStack(
						backgroundRect,
						container.NewVBox(
							widget.NewLabel(relPath),
							grid,
						),
					),
				}, mainPanel.Objects...)
			}
		}()
	}
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
	cmd := exec.Command("xdg-open", path)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error opening image with default app:", err)
	}
}

func newThumbnail(image *canvas.Image, path string) *ThumbnailWidget {
	t := &ThumbnailWidget{
		Image: image,
		OnTapped: func() {
			openImageWithDefaultApp(path)
		},
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *ThumbnailWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(t.Image)
}

func (t *ThumbnailWidget) Tapped(_ *fyne.PointEvent) {
	if t.OnTapped != nil {
		t.OnTapped()
	}
}
