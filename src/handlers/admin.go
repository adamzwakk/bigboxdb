package handlers

import (
	"archive/zip"
	"bytes"
    "io"
	"os"
    "log"
	"slices"
	"net/http"
	"path/filepath"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"

	// "github.com/adamzwakk/bigboxdb-server/db"
	// "github.com/adamzwakk/bigboxdb-server/models"
)

type ImportData struct{
	Title			string	`json:"title"`
	Description		string	`json:"description"`
	BoxType			int	`json:"box_type"`
	Width			float32	`json:"width"`
	Height			float32	`json:"height"`
	Depth			float32	`json:"depth"`
	Year			int	`json:"year"`
	Variant			string	`json:"variant"`
	Platform		string	`json:"platform"`
	ScanNotes		string	`json:"scan_notes,omitempty"`
	IGDBId			int		`json:"igdb_version,omitempty"`
	MobygamesId		int		`json:"mobygames_id,omitempty"`
	BBDBVersion		string	`json:"bbdb_version,omitempty"`
}

func AdminImport(c *gin.Context){
	//database := db.GetDB()

	destDir := "./uploads/scans/"
    os.MkdirAll(destDir, os.ModePerm)
	allowedFiles := []string{"back.webp", "bottom.webp", "box.glb", "box-low.glb", "front.webp", "info.json", "left.webp", "right.webp", "top.webp"}

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "No file uploaded")
		return
	}

	f, err := file.Open()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to open uploaded file")
		return
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, f); err != nil {
		c.String(http.StatusInternalServerError, "Failed to read uploaded file")
		return
	}

	zipReader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid zip file")
		return
	}

	// Info Process
	var jsonFile *zip.File
	for _, zf := range zipReader.File {
		if zf.Name == "info.json" {
			jsonFile = zf
            break
        }
	}
	if jsonFile == nil {
        c.String(http.StatusBadRequest, "JSON file not found in zip")
        return
    }
	rc, err := jsonFile.Open()
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to open JSON file")
        return
    }
    defer rc.Close()

	var data ImportData
    decoder := json.NewDecoder(rc)
    if err := decoder.Decode(&data); err != nil {
        c.String(http.StatusBadRequest, "Invalid JSON")
        return
    }

	slug := slug.Make(data.Title)
	c.String(http.StatusOK, slug)

	//c.JSON(http.StatusOK, data)
	return

	// Image Process
	for _, zf := range zipReader.File {
		if zf.Name == "info.json" {
			continue
		}
		if(!slices.Contains(allowedFiles, zf.Name)){
			c.String(http.StatusBadRequest, "Failed to read approve "+zf.Name)
			return
		}
		//log.Println("Processing:", zf.Name)
		rc, err := zf.Open()
		if err != nil {
			log.Println("Failed to open file in zip:", err)
			continue
		}

		outPath := filepath.Join(destDir, zf.Name)

		outFile, err := os.Create(outPath)
        if err != nil {
            log.Println("Failed to create file:", err)
            rc.Close()
            continue
        }

        // Copy contents
        _, err = io.Copy(outFile, rc)
        if err != nil {
            log.Println("Failed to copy file contents:", err)
        }
		rc.Close()
		outFile.Close()
	}

	c.String(http.StatusOK, "Zip processed successfully")
}