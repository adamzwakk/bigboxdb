package tools

import (
	"os"
	"fmt"
	"image"
	"strings"
	"path/filepath"
	"os/exec"
	"errors"

	_ "golang.org/x/image/tiff"
	"github.com/disintegration/imaging"
	"github.com/sunshineplan/imgconv"
)

type GameInfo struct {
	Title		string	 `json:"title"`
	Width       float32  `json:"width"`
	Height      float32  `json:"height"`
	Depth       float32  `json:"depth"`
	BoxType     uint      `json:"box_type"`
}

const (
	UpsizeRatio         = 80
	UpsizeRatioLow      = 60
)

func ProcessImage(srcPath string, dstPath, filename string, gWidth float32, gHeight float32, gDepth float32) error {
	fmt.Printf("Processing: %s\n", filename)
	ext := strings.ToLower(filepath.Ext(filename))

	if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Found webp, skipping...\n")
		return nil
	}

	img, err := imaging.Open(srcPath)
	if err != nil && (ext == ".tif" || ext == ".tiff") {
		fmt.Printf("TIFF decode failed, attempting ImageMagick conversion for %s\n", filename)
		
		pngPath := strings.TrimSuffix(srcPath, filepath.Ext(srcPath)) + ".png"
		defer os.Remove(pngPath)
		
		cmd := exec.Command("convert", srcPath, pngPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: ImageMagick conversion failed for %s - %v\n", filename, err)
			return nil
		}
		
		// Use the converted PNG instead
		srcPath = pngPath
		
		// Verify the conversion worked
		img, err = imaging.Open(srcPath)
		if err != nil {
			fmt.Printf("Warning: Converted image still invalid %s - %v\n", filename, err)
			return nil
		}
	}

	// Determine thumbnail size
	var width, height int
	if strings.HasPrefix(filename, "front") || strings.HasPrefix(filename, "back") ||
		strings.HasPrefix(filename, "gatefold_right") || strings.HasPrefix(filename, "gatefold_left") {
		width = int(gWidth * UpsizeRatio)
		height = int(gHeight * UpsizeRatio)
	} else if strings.HasPrefix(filename, "left") || strings.HasPrefix(filename, "right") {
		width = int(gDepth * UpsizeRatio)
		height = int(gHeight * UpsizeRatio)
	} else if strings.HasPrefix(filename, "top") || strings.HasPrefix(filename, "bottom") {
		width = int(gWidth * UpsizeRatio)
		height = int(gDepth * UpsizeRatio)
	}

	resized := imaging.Fit(img, width, height, imaging.Lanczos)

	// Save as WebP using cwebp command
	saveAsWebP(resized, dstPath)

	return nil
}

func saveAsWebP(img image.Image, path string) error {
	// Use imgconv to save directly as WebP
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Use imgconv to save as WebP with quality setting
	return imgconv.Write(outFile, img, &imgconv.FormatOption{Format: imgconv.WEBP})
}