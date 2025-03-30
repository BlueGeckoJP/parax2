package main

import (
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func OpenSettingsWindow() {
	settingsWindow := myApp.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(400, 400))

	thumbnailSizeEntry := widget.NewEntry()
	thumbnailSizeEntry.SetText(getTSizeString(thumbnailWidth, thumbnailHeight))
	thumbnailSizeEntry.SetPlaceHolder("Thumbnail Size (e.g. 200x200)")
	thumbnailSizeEntry.Validator = func(text string) error {
		_, _, err := getTSizeFromString(text)
		if err != nil {
			return err
		}
		return nil
	}
	thumbnailSizeEntry.OnChanged = func(text string) {
		w, h, err := getTSizeFromString(text)
		if err == nil {
			thumbnailWidth = w
			thumbnailHeight = h
		}
	}

	thumbnailSizeContainer := container.NewBorder(
		nil,
		nil,
		widget.NewLabel("Thumbnail Size:"),
		nil,
		thumbnailSizeEntry,
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

func getTSizeString(w, h float32) string {
	return strconv.Itoa(int(w)) + "x" + strconv.Itoa(int(h))
}

func getTSizeFromString(s string) (float32, float32, error) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid format")
	}
	w, err := strconv.ParseFloat(parts[0], 32)
	if err != nil {
		return 0, 0, err
	}
	h, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return 0, 0, err
	}
	return float32(w), float32(h), nil
}
