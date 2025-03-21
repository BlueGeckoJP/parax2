package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func OpenSettingsWindow() {
	settingsWindow := myApp.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(400, 400))

	thumbnailSizeWEntry := widget.NewEntry()
	thumbnailSizeWEntry.SetText(strconv.Itoa(int(thumbnailWidth)))
	thumbnailSizeWEntry.OnChanged = func(text string) {
		w, err := strconv.ParseFloat(text, 32)
		if err == nil {
			thumbnailWidth = float32(w)
		}
	}

	thumbnailSizeHEntry := widget.NewEntry()
	thumbnailSizeHEntry.SetText(strconv.Itoa(int(thumbnailHeight)))
	thumbnailSizeHEntry.OnChanged = func(text string) {
		h, err := strconv.ParseFloat(text, 32)
		if err == nil {
			thumbnailHeight = float32(h)
		}
	}

	thumbnailSizeContainer := container.NewHBox(
		widget.NewLabel("Thumbnail Size:"),
		thumbnailSizeWEntry,
		thumbnailSizeHEntry,
	)

	viewModeSelect := widget.NewSelect([]string{"List", "Grid"}, func(selected string) {
		if selected == "List" {
			mainPanel.viewMode = 0
		} else if selected == "Grid" {
			mainPanel.viewMode = 1
		}
	})
	if mainPanel.viewMode == 0 {
		viewModeSelect.Selected = "List"
	} else if mainPanel.viewMode == 1 {
		viewModeSelect.Selected = "Grid"
	}
	viewModeSelect.OnChanged = func(selected string) {
		if selected == "List" {
			mainPanel.viewMode = 0
		} else if selected == "Grid" {
			mainPanel.viewMode = 1
		}
	}

	viewModeContainer := container.NewHBox(
		widget.NewLabel("View Mode:"),
		viewModeSelect,
	)

	refreshButton := widget.NewButton("Refresh", func() {
		mainPanel.Update(mainPanel.originalPath)
	})

	topContainer := container.NewVBox(
		thumbnailSizeContainer,
		viewModeContainer,
		refreshButton,
	)

	settingsWindow.SetContent(topContainer)
	settingsWindow.Show()
}
