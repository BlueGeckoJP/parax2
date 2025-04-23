package main

import (
	"image"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"weak"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

type Entries struct {
	Path   string
	Images []*ImageEntry
}

type ImageEntry struct {
	Path  string
	Image weak.Pointer[canvas.Image]
}

var supportedExtensions = regexp.MustCompile(`.jpg|.jpeg|.png|.webp`)

func search(root string, maxDepth int) []*Entries {
	result := make(map[string]*Entries)
	baseDepth := getDepth(root)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if getDepth(path)-1 > baseDepth+maxDepth {
			return filepath.SkipDir
		}

		if d.IsDir() {
			return nil
		}

		dir := filepath.Dir(path)
		entries, ok := result[dir]
		if !ok {
			e := &Entries{
				Path:   dir,
				Images: []*ImageEntry{},
			}
			result[dir] = e
			entries = e
		}

		if IsSupportedExtension(path) {
			entries.Images = append(entries.Images, &ImageEntry{Path: path})
		}

		return nil
	})

	entries := make([]*Entries, 0, len(result))

	for _, entry := range result {
		if len(entry.Images) > 0 {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (e *Entries) LoadAll() {
	for _, i := range e.Images {
		err := e.Load(i)
		if err != nil {
			log.Println("An error occurred while LoadAll function:", err)
		}
	}
}

func (e *Entries) Load(i *ImageEntry) error {
	if i.Image.Value() == nil {
		f, err := os.Open(i.Path)
		if err != nil {
			return err
		}
		defer f.Close()

		img, err := getScaled(f)
		if err != nil {
			return err
		}

		canvasImage := canvas.NewImageFromImage(img)
		canvasImage.FillMode = canvas.ImageFillContain
		canvasImage.SetMinSize(fyne.NewSize(thumbnailWidth, thumbnailHeight))

		i.Image = weak.Make(canvasImage)
	}

	return nil
}

func (e *Entries) Get(i *ImageEntry) *canvas.Image {
	img := i.Image.Value()
	if img != nil {
		return img
	} else {
		e.Load(i)
		return i.Image.Value()
	}
}

func getScaled(f *os.File) (image.Image, error) {
	var img image.Image
	var err error

	if filepath.Ext(f.Name()) == ".webp" {
		img, err = webp.Decode(f)
		if err != nil {
			return nil, err
		}
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

func getDepth(path string) int {
	return strings.Count(path, string(os.PathSeparator))
}

func IsSupportedExtension(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return supportedExtensions.MatchString(ext)
}
