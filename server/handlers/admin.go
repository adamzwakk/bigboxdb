package handlers

import (
	"archive/zip"
	"bytes"
    "io"
	"os"
    "log"
	"strings"
	"fmt"
	"slices"
	"strconv"
	"net/http"
	"path/filepath"
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"gorm.io/gorm/clause"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
	"github.com/adamzwakk/bigboxdb-server/tools"
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

var destDir = "./uploads/scans/"
// var allowedFiles = []string{"back.webp", "bottom.webp", "box.glb", "box-low.glb", "front.webp", "info.json", "left.webp", "right.webp", "top.webp"}
var allowedFiles = []string{
	"back.tif", 
	"bottom.tif", 
	"box.glb",
	"box-low.glb", 
	"front.tif", 
	"info.json", 
	"left.tif", 
	"right.tif", 
	"top.tif", 
	"gatefold_left.tif", 
	"gatefold_right.tif",
	"back.webp", 
	"bottom.webp", 
	"front.webp", 
	"left.webp", 
	"right.webp", 
	"top.webp", 
	"gatefold_left.webp", 
	"gatefold_right.webp",
}

// Testing curl - curl -H "Authorization: Bearer {some key}" -X PUT http://localhost:8080/api/admin/import -F "file=@./testbox.zip" -H "Content-Type: multipart/form-data"
func AdminImport(c *gin.Context){
	file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    
    uploadedFile, err := file.Open()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
        return
    }
    defer uploadedFile.Close()
    
    zipData, err := io.ReadAll(uploadedFile)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
        return
    }
    
    if err := ImportZip(zipData); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "OK"})

	return
}

type FileSource interface {
	ReadJSON(filename string) ([]byte, error)
	ListFiles() ([]string, error)
	GetFilePath(filename string) (string, bool, error) // returns path, isTemp, error
}

// ZipSource implements FileSource for zip files
type ZipSource struct {
	reader *zip.Reader
}

func (z *ZipSource) ReadJSON(filename string) ([]byte, error) {
	for _, f := range z.reader.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("file not found: %s", filename)
}

func (z *ZipSource) ListFiles() ([]string, error) {
	var files []string
	for _, f := range z.reader.File {
		if !f.FileInfo().IsDir() {
			files = append(files, f.Name)
		}
	}
	return files, nil
}

func (z *ZipSource) GetFilePath(filename string) (string, bool, error) {
	for _, f := range z.reader.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return "", false, err
			}
			defer rc.Close()

			tmpFile, err := os.CreateTemp("", "zipimg-*"+filepath.Ext(filename))
			if err != nil {
				return "", false, err
			}
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, rc); err != nil {
				os.Remove(tmpFile.Name())
				return "", false, err
			}

			return tmpFile.Name(), true, nil // true = is temporary
		}
	}
	return "", false, fmt.Errorf("file not found: %s", filename)
}

// DirectorySource implements FileSource for local directories
type DirectorySource struct {
	path string
}

func (d *DirectorySource) ReadJSON(filename string) ([]byte, error) {
	return os.ReadFile(filepath.Join(d.path, filename))
}

func (d *DirectorySource) ListFiles() ([]string, error) {
	entries, err := os.ReadDir(d.path)
	if err != nil {
		return nil, err
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}
	return files, nil
}

func (d *DirectorySource) GetFilePath(filename string) (string, bool, error) {
	path := filepath.Join(d.path, filename)
	if _, err := os.Stat(path); err != nil {
		return "", false, err
	}
	return path, false, nil // false = not temporary, don't delete
}

// Main import function - works with both sources
func ImportFromSource(source FileSource) error {
	// Read and parse JSON
	jsonData, err := source.ReadJSON("info.json")
	if err != nil {
		return fmt.Errorf("JSON file not found: %w", err)
	}

	var data ImportData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	database := db.GetDB()
	slugTitle := slug.Make(data.Title)
	variantDesc := data.Variant

	if data.BBDBVersion == nil {
		data.BoxType++
	}

	userName := os.Getenv("BBDB_ADMIN_NAME")
	if data.ContributedBy != nil {
		userName = *data.ContributedBy
	}

	worthFront := true
	if data.WorthFrontView != nil {
		worthFront = *data.WorthFrontView
	}

	var user models.User
	if err := database.FirstOrCreate(&user, models.User{Name: userName}).Error; err != nil {
		return fmt.Errorf("could not find/create User")
	}

	var platform models.Platform
	if err := database.FirstOrCreate(&platform, models.Platform{Name: data.Platform, Slug: slug.Make(data.Platform)}).Error; err != nil {
		return fmt.Errorf("could not find/create Platform")
	}

	var dev models.Developer
	database.Where(models.Developer{Name: string(data.Developer)}).Assign(models.Developer{Slug: slug.Make(string(data.Developer))}).FirstOrCreate(&dev)

	var pub models.Publisher
	database.Where(models.Publisher{Name: string(data.Publisher)}).Assign(models.Publisher{Slug: slug.Make(string(data.Publisher))}).FirstOrCreate(&pub)

	game := models.Game{
		Title:       data.Title,
		Slug:        slugTitle,
		Description: data.Description,
		Year:        data.Year,
		PlatformID:  platform.ID,
		Variants: []models.Variant{
			{
				BoxTypeID:      data.BoxType,
				Description:    variantDesc,
				Slug:           slug.Make(fmt.Sprintf("%s-%d", variantDesc, data.BoxType)),
				DeveloperID:    dev.ID,
				PublisherID:    pub.ID,
				Width:          data.Width,
				Height:         data.Height,
				Depth:          data.Depth,
				WorthFrontView: worthFront,
				UserID:         user.ID,
			},
		},
	}

	database.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title",
			"description",
		}),
	}).Create(&game)

	variantID := game.Variants[0].ID
	if variantID == 0 {
		var variant models.Variant
		database.Where("slug = ?", game.Variants[0].Slug).First(&variant)
		variantID = variant.ID
	}

	// Process images
	var texPaths []string
	wd, err := os.Getwd()
	gameDir := filepath.Join(wd, "uploads/scans", slugTitle, strconv.Itoa(int(variantID)))
	os.MkdirAll(gameDir, os.ModePerm)

	files, err := source.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	for _, filename := range files {
		if filename == "info.json" {
			continue
		}

		if !slices.Contains(allowedFiles, filename) {
			return fmt.Errorf("failed to approve " + filename)
		}

		srcPath, isTemp, err := source.GetFilePath(filename)
		if err != nil {
			return fmt.Errorf("failed to get file: %w", err)
		}
		if isTemp {
			defer os.Remove(srcPath)
		}

		dstPath := gameDir + "/" + filename
		dstPath = strings.ReplaceAll(dstPath, ".tif", ".webp")

		if err := tools.ProcessImage(srcPath, dstPath, filename, data.Width, data.Height, data.Depth); err != nil {
			return fmt.Errorf("failed to process image: " + err.Error())
		}

		texPaths = append(texPaths, dstPath)
	}

	gameInfo := &tools.GameInfo{
		Title:   data.Title,
		Width:   data.Width,
		Height:  data.Height,
		Depth:   data.Depth,
		BoxType: data.BoxType,
	}

	if os.Getenv("APP_ENV") != "production" {
		log.Println("Making glb file")
	}
	if err := tools.GenerateGLTFBox(gameInfo, texPaths, gameDir, false); err != nil {
		return fmt.Errorf("failed to process glb file: " + err.Error())
	}
	if os.Getenv("APP_ENV") != "production" {
		log.Println("Making low glb file")
	}
	if err := tools.GenerateGLTFBox(gameInfo, texPaths, gameDir, true); err != nil {
		return fmt.Errorf("failed to process glb file: " + err.Error())
	}

	tools.CleanupKTX2Files(gameDir)

	return nil
}

func ImportZip(zipData []byte) error {
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("invalid zip file: %w", err)
	}
	return ImportFromSource(&ZipSource{reader: reader})
}

func ImportDirectory(dirPath string) error {
	return ImportFromSource(&DirectorySource{path: dirPath})
}

func ImportLocal(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if info.IsDir() {
		return ImportDirectory(path)
	} else {
		zipData, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read zip: %w", err)
		}
		return ImportZip(zipData)
	}
}