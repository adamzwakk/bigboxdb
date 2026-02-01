package handlers

import (
	"archive/zip"
	"bytes"
    "io"
	"os"
    "log"
	"fmt"
	"slices"
	"net/http"
	"path/filepath"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"gorm.io/gorm/clause"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

type ImportData struct{
	Title			string	`json:"title"`
	Description		*string	`json:"description,omitempty"`
	SeriesSort		string	`json:"series"`
	BoxType			uint	`json:"box_type"`
	Width			float32	`json:"width"`
	Height			float32	`json:"height"`
	Depth			float32	`json:"depth"`
	Year			int	`json:"year"`
	Variant			string	`json:"variant"`
	Developer		FirstString	`json:"developer"`
	Publisher		FirstString	`json:"publisher"`
	Platform		string	`json:"platform"`
	ScanNotes		string	`json:"scan_notes,omitempty"`
	IGDBId			int		`json:"igdb_version,omitempty"`
	MobygamesId		int		`json:"mobygames_id,omitempty"`
	BBDBVersion		*int	`json:"bbdb_version,omitempty"`
	ContributedBy	*string	`json:"contributed_by,omitempty"`
	WorthFrontView	*bool	`json:"worth_front_view,omitempty"`
}

type FirstString string
func (f *FirstString) UnmarshalJSON(data []byte) error {
    // Try to unmarshal as string
    var s string
    if err := json.Unmarshal(data, &s); err == nil {
        *f = FirstString(s)
        return nil
    }

    // Try to unmarshal as []string
    var arr []string
    if err := json.Unmarshal(data, &arr); err == nil {
        if len(arr) > 0 {
            *f = FirstString(arr[0])
        } else {
            *f = ""
        }
        return nil
    }

    // If neither, set empty
    *f = ""
    return nil
}

// Testing curl - curl -H "Authorization: Bearer {some key}" -X PUT http://localhost:8080/api/admin/import -F "file=@./testbox.zip" -H "Content-Type: multipart/form-data"
func AdminImport(c *gin.Context){
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

	database := db.GetDB()
	slugTitle := slug.Make(data.Title)
	variantDesc := data.Variant

	if data.BBDBVersion == nil {
		data.BoxType++			// old boxes started at 0, whoops
	}

	userName := "UncleHans" // default
	if data.ContributedBy != nil {
		userName = *data.ContributedBy
	}

	worthFront := true
	if data.WorthFrontView != nil {
		worthFront = *data.WorthFrontView
	}

	var user models.User
	if err := database.FirstOrCreate(&user, models.User{Name: userName}).Error; err != nil {
		c.String(http.StatusInternalServerError, "Could not find/create User")
	}

	var platform models.Platform
	if err := database.FirstOrCreate(&platform, models.Platform{Name: data.Platform, Slug: slug.Make(data.Platform)}).Error; err != nil {
		c.String(http.StatusInternalServerError, "Could not find/create Platform")
	}

	var dev models.Developer
	database.Where(models.Developer{Name: string(data.Developer)}).Assign(models.Developer{Slug: slug.Make(string(data.Developer))}).FirstOrCreate(&dev)

	var pub models.Publisher
	database.Where(models.Publisher{Name: string(data.Publisher)}).Assign(models.Publisher{Slug: slug.Make(string(data.Publisher))}).FirstOrCreate(&pub)

	game := models.Game{
		Title:			data.Title,
		Slug:			slugTitle,
		Description: 	data.Description,
		Year:			data.Year,
		PlatformID:		platform.ID,
		Variants:		[]models.Variant{
			{
				BoxTypeID:	data.BoxType,
				Description: variantDesc,
				Slug:		slug.Make(fmt.Sprintf("%s-%d", variantDesc, data.BoxType)),

				DeveloperID:	dev.ID,
				PublisherID:	pub.ID,

				Width:		data.Width,
				Height:		data.Height,
				Depth:		data.Depth,
				WorthFrontView:	worthFront,

				UserID:		user.ID,
			},
		},
	}

	database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title",
			"description",
		}),
	}).Create(&game)

	c.String(http.StatusOK, slugTitle)

	//c.JSON(http.StatusOK, data)
	return

	// Image Process
	for _, zf := range zipReader.File {
		if zf.Name == "info.json" {
			continue // we already processed you!
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