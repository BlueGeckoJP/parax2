package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"net/http"
	_ "net/http/pprof"
)

type Entry struct {
	Path     string
	Children []*Entry
	Depth    int
	IsDir    bool
}

var myApp fyne.App
var mainPanel *MainPanel

var thumbnailWidth float32 = 200
var thumbnailHeight float32 = 200
var maxDepth = 2
var wgMax = 8

var thumbnailCache = NewLRUCache(5000)
var config = loadConfig()

var directoryTree *widget.Tree
var directoryTreeLabel *widget.Label
var myWindow fyne.Window

func main() {
	ifDebug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if *ifDebug {
		addr := "localhost:6060"
		log.Println("Enabled debug mode!! :", addr)
		go func() {
			err := http.ListenAndServe(addr, nil)
			log.Println(err)
		}()
	}

	myApp = app.New()
	myWindow = myApp.NewWindow("parax2")
	mainPanel = newMainPanel()

	if config != nil {
		if config.ViewMode == 0 || config.ViewMode == 1 {
			mainPanel.viewMode = config.ViewMode
		}
		if config.MaxDepth > 0 {
			maxDepth = config.MaxDepth
		}
		if config.CacheLimit > 0 {
			thumbnailCache = NewLRUCache(config.CacheLimit)
		}
		if config.WGMax > 1 {
			wgMax = config.WGMax
		}
		log.Println("Received raw config: ", config)
	}

	directoryTree = widget.NewTree(
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			if id == "" {
				children := make([]widget.TreeNodeID, 0)
				for _, entry := range mainPanel.entries {
					children = append(children, entry.Path)
				}
				return children
			}

			for _, entry := range mainPanel.entries {
				if entry.Path == id && entry.IsDir {
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
				if entry.Path == id && !entry.IsDir {
					openImageWithDefaultApp(entry.Path)
				}
				if entry.Children != nil {
					findId(entry.Children)
				}
			}
		}
		go findId(mainPanel.entries)
	}

	directoryTreeLabel = widget.NewLabel("Tree in " + filepath.Base(mainPanel.originalPath))

	leftPanel := container.New(layout.NewBorderLayout(directoryTreeLabel, nil, nil, nil), directoryTreeLabel, directoryTree)

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Open Folder", func() {
				dialog.ShowFolderOpen(func(reader fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, myWindow)
						return
					}
					if reader != nil {
						mainPanel.Update(reader.Path())
					}
				}, myWindow)
			}),
			fyne.NewMenuItem("Open Settings", func() {
				OpenSettingsWindow()
			}),
		),
	)

	myWindow.SetMainMenu(mainMenu)

	split := container.NewHSplit(leftPanel, container.NewVScroll(mainPanel.c))
	split.SetOffset(0.2)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
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
		log.Println("Error starting command: ", err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Println("Error waiting for command: ", err)
		return
	}
}
