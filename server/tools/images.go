package tools

import (
	"os"
	"fmt"
	"image"
	"strings"
	"path/filepath"
	"os/exec"
	"errors"

	"github.com/disintegration/imaging"
	"github.com/sunshineplan/imgconv"
)

const (
	UpsizeRatio         = 80
	UpsizeRatioLow      = 60
)

func ProcessImage(srcPath string, dstPath, filename string, gWidth float32, gHeight float32, gDepth float32) error {
	if os.Getenv("APP_ENV") != "production" {
		fmt.Printf("Processing: %s\n", filename)
	}

	if _, err := os.Stat(dstPath); !errors.Is(err, os.ErrNotExist) {
        return nil // Already exists
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

	processImageWithVips(srcPath, dstPath, width, height)

	return nil
}

func saveAsWebP(img image.Image, path string) error {
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return imgconv.Write(outFile, img, &imgconv.FormatOption{Format: imgconv.WEBP})
}

func processImageWithVips(srcPath, dstPath string, width, height int) error {
    ext := strings.ToLower(filepath.Ext(srcPath))
    
    if ext == ".tif" || ext == ".tiff" {
        cmd := exec.Command("vipsthumbnail", srcPath,
            "-o", dstPath+"[Q=80]",
            "-s", fmt.Sprintf("%dx%d", width, height),
        )
        return cmd.Run()
    }
    
    // For non-TIFF, use your existing imaging pipeline
    img, err := imaging.Open(srcPath)
    if err != nil {
        return err
    }
    resized := imaging.Fit(img, width, height, imaging.Lanczos)
    return saveAsWebP(resized, dstPath)
}