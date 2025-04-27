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
)

const (
	ViewModeList = iota
	ViewModeGrid
)

type PathID = string

var backgroundRect = canvas.NewRectangle(color.Color(color.RGBA{R: 51, G: 51, B: 51, A: 255}))

func newMainPanel() *MainPanel {
	return &MainPanel{
		c:            container.NewVBox(),
		viewMode:     ViewModeGrid,
		entries:      []*Entries{},
		originalPath: ".",
		containerMap: make(map[PathID]*CustomContainer),
	}
}

type MainPanel struct {
	c            *fyne.Container
	viewMode     int
	entries      []*Entries
	originalPath string
	containerMap map[string]*CustomContainer
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

	m.entries, err = search(currentPath, maxDepth)
	if err != nil {
		log.Println("An error occurred while searching entries:", err)
	}
	log.Println("Loaded", len(m.entries), "entries")

	m.c.Objects = nil

	myWindow.SetTitle("Loading - parax2")

	m.originalPath = currentPath

	m.containerMap = make(map[string]*CustomContainer)

	for _, entry := range m.entries {
		switch m.viewMode {
		case ViewModeList:
			l := newListContainer(getRelPath(m.originalPath, entry.Path))
			m.c.Add(l)
			var lcc CustomContainer = l
			m.containerMap[entry.Path] = &lcc
		case ViewModeGrid:
			g := newGridContainer(getRelPath(m.originalPath, entry.Path))
			m.c.Add(g)
			var gcc CustomContainer = g
			m.containerMap[entry.Path] = &gcc
		}
	}

	m.sortContainers()

	err = m.loadAllImages()
	if err != nil {
		log.Println("An error occurred while loading all images:", err)
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
		return errors.New("the container specified by pathId could not be found")
	}

	if len(m.entries) == 0 {
		return errors.New("entries list is empty")
	}

	var entries *Entries
	for _, e := range m.entries {
		if e.Path == pathId {
			entries = e
			break
		}
	}
	if entries == nil {
		return errors.New("entries not found")
	}
	entries.LoadAll()

	wg := newWGC()

	for _, i := range entries.Images {
		wg.Add(func() {
			defer wg.Done()
			img, err := entries.Get(i)
			if err != nil {
				log.Println("An error occurred while get image from entries:", err)
				return
			}
			if img != nil {
				thumbnail := newThumbnail(img, i.Path)
				(*c).Add(thumbnail)
			} else {
				log.Println("The image in the entry is null.")
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
			return m.c.Objects[i].(*ListContainer).Title() < m.c.Objects[j].(*ListContainer).Title()
		})
	} else if m.viewMode == ViewModeGrid {
		sort.SliceStable(m.c.Objects, func(i, j int) bool {
			return m.c.Objects[i].(*GridContainer).Title() < m.c.Objects[j].(*GridContainer).Title()
		})
	}
}

func sortObjects(c *CustomContainer) {
	sort.Slice((*c).Objects(), func(i, j int) bool {
		reAll := regexp.MustCompile(`(\d+)|(\D+)`)
		reNumPerfect := regexp.MustCompile(`^\d+$`)

		iPath := (*c).Objects()[i].(*ThumbnailWidget).Path
		jPath := (*c).Objects()[j].(*ThumbnailWidget).Path
		partsI := reAll.FindAllString(iPath, -1)
		partsJ := reAll.FindAllString(jPath, -1)

		for n := range min(len(partsI), len(partsJ)) {
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
	rel, err := filepath.Rel(baseDirPath, path)
	if err != nil {
		return path
	}
	return rel
}
