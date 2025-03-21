package main

import (
	"image"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

const (
	ViewModeList = iota
	ViewModeGrid
)

var supportedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

func newMainPanel() *MainPanel {
	return &MainPanel{
		c:            container.NewVBox(),
		viewMode:     ViewModeGrid,
		entries:      []*Entry{},
		originalPath: ".",
	}
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedExtensions[ext]
}

func loadThumbnail(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var img image.Image

	if filepath.Ext(path) == ".webp" {
		img, err = webp.Decode(f)
		if err != nil {
			return nil, err
		}
		return img, nil
	} else {
		img, _, err = image.Decode(f)
		if err != nil {
			return nil, err
		}
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width > height {
		height = (height * int(thumbnailWidth)) / width
		width = int(thumbnailWidth)
	} else {
		width = (width * int(thumbnailHeight)) / height
		height = int(thumbnailHeight)
	}

	scaledSize := image.Rect(0, 0, width, height)
	scaled := image.NewRGBA(scaledSize)
	draw.BiLinear.Scale(scaled, scaledSize, img, img.Bounds(), draw.Over, nil)

	return scaled, nil
}

type MainPanel struct {
	c            *fyne.Container
	viewMode     int
	entries      []*Entry
	originalPath string
}

func (m *MainPanel) Update(currentPath string) {
	if m.originalPath == "." && m.originalPath == currentPath {
		return
	}

	log.Println("MainPanel.Update called with path:", currentPath)

	currentPath, err := filepath.Abs(currentPath)
	if err != nil {
		log.Println("Error getting absolute path:", err)
		return
	}

	m.c.Objects = nil
	m.entries = nil

	myWindow.SetTitle("Loading - parax2")

	m.originalPath = currentPath
	m.update(currentPath, 0, &m.entries)

	m.sortContainers()

	if directoryTreeLabel != nil {
		directoryTreeLabel.SetText("Tree in " + filepath.Base(m.originalPath))
	}
	if directoryTree != nil {
		directoryTree.Refresh()
	}

	myWindow.SetTitle("parax2")
	log.Println("MainPanel.Update done")
}

func (m *MainPanel) update(currentPath string, depth int, entries *[]*Entry) {
	var c *fyne.Container
	switch m.viewMode {
	case ViewModeList:
		c = container.NewHBox()
	case ViewModeGrid:
		c = container.NewGridWrap(fyne.NewSize(thumbnailWidth, thumbnailHeight))
	}

	f, err := os.Open(currentPath)
	if err != nil {
		log.Println("Error opening directory:", err)
		return
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		log.Println("Error reading directory:", err)
		return
	}

	wg := &WGWithCounter{
		wg:    sync.WaitGroup{},
		count: 0,
		max:   wgMax,
	}

	for _, v := range files {
		p := filepath.Join(currentPath, v.Name())
		if isImageFile(v.Name()) {
			wg.Add(1, func() {
				defer wg.Done()
				var thumbnail *ThumbnailWidget
				image, exists := thumbnailCache.get(p)
				if !exists {
					thumbnailImage, err := loadThumbnail(p)
					if err != nil {
						log.Println("Error loading thumbnail:", err)
						return
					}

					canvasImage := canvas.NewImageFromImage(thumbnailImage)
					canvasImage.FillMode = canvas.ImageFillContain
					canvasImage.SetMinSize(fyne.NewSize(thumbnailWidth, thumbnailHeight))

					thumbnail = newThumbnail(canvasImage, p)

					thumbnailCache.add(p, canvasImage)
				} else {
					thumbnail = newThumbnail(image, p)
				}
				c.Add(thumbnail)
				*entries = append(*entries, &Entry{
					Path:     p,
					Children: nil,
					IsDir:    false,
					Depth:    depth,
				})
			})
		} else if v.IsDir() && v.Name()[0] != '.' {
			if maxDepth > depth {
				entry := &Entry{
					Path:     p,
					Children: []*Entry{},
					IsDir:    true,
					Depth:    depth,
				}
				m.update(p, depth+1, &entry.Children)
				*entries = append(*entries, entry)
			}
		}
	}

	wg.wg.Wait()

	if c.Objects != nil {
		sort.SliceStable(c.Objects, func(i, j int) bool {
			return c.Objects[i].(*ThumbnailWidget).Path < c.Objects[j].(*ThumbnailWidget).Path
		})

		relPath, _ := filepath.Rel(m.originalPath, currentPath)

		var cVBox *fyne.Container
		if m.viewMode == ViewModeList {
			cVBox = container.NewVBox(
				widget.NewLabel(relPath),
				container.NewHScroll(c),
			)
		} else {
			cVBox = container.NewVBox(
				widget.NewAccordion(
					widget.NewAccordionItem(
						relPath,
						c,
					),
				),
			)
		}

		backgroundRect := canvas.NewRectangle(color.Color(color.RGBA{51, 51, 51, 255}))

		m.c.Add(container.NewStack(backgroundRect, cVBox))
	}
}

func (m *MainPanel) sortContainers() {
	if m.viewMode == ViewModeList {
		sort.SliceStable(m.c.Objects, func(i, j int) bool {
			return m.c.Objects[i].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Label).Text < m.c.Objects[j].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Label).Text
		})
	} else if m.viewMode == ViewModeGrid {
		sort.SliceStable(m.c.Objects, func(i, j int) bool {
			return m.c.Objects[i].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Accordion).Items[0].Title < m.c.Objects[j].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*widget.Accordion).Items[0].Title
		})
	}
}
