package main

import (
	"errors"
	"image/color"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	ViewModeList = iota
	ViewModeGrid
)

type PathID = string

func newMainPanel() *MainPanel {
	return &MainPanel{
		c:            container.NewVBox(),
		viewMode:     ViewModeGrid,
		entries:      []*Entries{},
		originalPath: ".",
		containerMap: make(map[PathID]*fyne.Container),
	}
}

type MainPanel struct {
	c            *fyne.Container
	viewMode     int
	entries      []*Entries
	originalPath string
	containerMap map[string]*fyne.Container
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

	m.entries = search(currentPath, maxDepth)
	log.Println("Loaded", len(m.entries), "entries")

	m.c.Objects = nil

	myWindow.SetTitle("Loading - parax2")

	m.originalPath = currentPath

	backgroundRect := canvas.NewRectangle(color.Color(color.RGBA{51, 51, 51, 255}))

	for _, entry := range m.entries {
		var outer *fyne.Container
		switch m.viewMode {
		case ViewModeList:
			c := container.NewHBox()
			m.containerMap[entry.Path] = c
			rel := getRelPath(m.originalPath, entry.Path)
			outer = container.NewVBox(
				widget.NewLabel(rel), container.NewHScroll(c),
			)
		case ViewModeGrid:
			c := container.NewGridWrap(fyne.NewSize(thumbnailWidth, thumbnailWidth))
			m.containerMap[entry.Path] = c
			rel := getRelPath(m.originalPath, entry.Path)
			outer = container.NewVBox(widget.NewAccordion(widget.NewAccordionItem(rel, c)))
		}

		m.c.Add(container.NewStack(backgroundRect, outer))
	}

	m.sortContainers()

	err = m.loadAllImages()
	if err != nil {
		println(err)
	}

	if directoryTreeLabel != nil {
		directoryTreeLabel.SetText("Tree in " + filepath.Base(m.originalPath))
	}
	if directoryTree != nil {
		directoryTree.Refresh()
	}

	myWindow.SetTitle("parax2")
	log.Println("MainPanel.Update done")
}

func (m *MainPanel) loadImages(pathId PathID) error {
	c := m.containerMap[pathId]
	if c == nil {
		return errors.New("The container specified by pathId could not be found.")
	}

	if len(m.entries) == 0 {
		return errors.New("Entries list is empty.")
	}

	var entries *Entries
	for _, e := range m.entries {
		if e.Path == pathId {
			entries = e
			break
		}
	}
	if entries == nil {
		return errors.New("Entries not found.")
	}
	entries.LoadAll()

	wg := newWGWC()

	for _, i := range entries.Images {
		wg.Add(func() {
			defer wg.Done()
			img := entries.Get(i)
			if img != nil {
				thumbnail := newThumbnail(img, i.Path)
				c.Add(thumbnail)
			} else {
				println("not found", i.Path)
			}
		})
	}

	wg.wg.Wait()

	sortObjects(c)

	return nil
}

func (m *MainPanel) loadAllImages() error {
	for key := range m.containerMap {
		err := m.loadImages(key)
		if err != nil {
			return err
		}
	}

	return nil
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

func sortObjects(c *fyne.Container) {
	sort.Slice(c.Objects, func(i, j int) bool {
		reAll := regexp.MustCompile(`(\d+)|(\D+)`)
		reNumPerfect := regexp.MustCompile(`^\d+$`)

		iPath := c.Objects[i].(*ThumbnailWidget).Path
		jPath := c.Objects[j].(*ThumbnailWidget).Path
		partsI := reAll.FindAllString(iPath, -1)
		partsJ := reAll.FindAllString(jPath, -1)

		for n := range max(len(partsI), len(partsJ)) {
			partI := partsI[n]
			partJ := partsJ[n]

			switch {
			case n >= len(partsI):
				return false
			case n >= len(partsJ):
				return true
			}

			if partI == partJ {
				continue
			}

			if reNumPerfect.MatchString(partI) {
				if reNumPerfect.MatchString(partJ) {
					numI, _ := strconv.Atoi(partI)
					numJ, _ := strconv.Atoi(partJ)
					return numI < numJ
				} else {
					return true
				}
			} else {
				if reNumPerfect.MatchString(partJ) {
					return false
				} else {
					return partI < partJ
				}
			}
		}

		return false
	})
}

func getRelPath(baseDirPath string, path string) string {
	println(baseDirPath, path)
	rel, err := filepath.Rel(baseDirPath, path)
	if err != nil {
		return path
	}
	return rel
}
