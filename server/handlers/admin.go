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
	"github.com/meilisearch/meilisearch-go"
	"github.com/dchest/uniuri"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
	"github.com/adamzwakk/bigboxdb-server/tools"
)

type ImportData struct{
	Title			string	`json:"title"`
	Description		*string	`json:"description,omitempty"`
	Region			*string `json:"region,omitempty"`
	SeriesSort		string	`json:"series"`
	BoxType			uint	`json:"box_type"`
	Width			float32	`json:"width"`
	Height			float32	`json:"height"`
	GatefoldTransparent		*bool `json:"gatefold_transparent"`
	Depth			float32	`json:"depth"`
	Year			int	`json:"year"`
	Variant			string	`json:"variant"`
	Developer		FirstString	`json:"developer"`
	Publisher		FirstString	`json:"publisher"`
	Platform		string	`json:"platform"`
	ScanNotes		string	`json:"scan_notes,omitempty"`
	IGDBId			*int		`json:"igdb_version,omitempty"`
	MobygamesId		*int		`json:"mobygames_id,omitempty"`
	BBDBVersion		*int	`json:"bbdb_version,omitempty"`
	ContributedBy	*string	`json:"contributed_by,omitempty"`
	Links			map[string]string `json:"links"`
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
	"gatefold_back_left.tif", 
	"gatefold_back_right.tif",
	"gatefold_front_left.tif", 
	"gatefold_front_right.tif",
	"back.webp", 
	"bottom.webp", 
	"front.webp", 
	"left.webp", 
	"right.webp", 
	"top.webp", 
	"gatefold_left.webp", 
	"gatefold_right.webp",
	"gatefold_back_left.webp", 
	"gatefold_back_right.webp",
	"gatefold_front_left.webp", 
	"gatefold_front_right.webp",
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
	meili := db.InitMeiliSearch()
	slugTitle := slug.Make(data.Title)
	variantDesc := data.Variant

	if data.BBDBVersion == nil {
		data.BoxType++
	}

	userName := os.Getenv("BBDB_ADMIN_NAME")
	if data.ContributedBy != nil {
		userName = *data.ContributedBy
	}

	var user models.User
	if err := database.FirstOrCreate(&user, models.User{Name: userName, ApiKey: uniuri.NewLen(24)}).Error; err != nil {
		return fmt.Errorf("could not find/create User")
	}

	var platform models.Platform
	if err := database.FirstOrCreate(&platform, models.Platform{Name: data.Platform, Slug: slug.Make(data.Platform)}).Error; err != nil {
		return fmt.Errorf("could not find/create Platform")
	}

	regString := "US"
	if data.Region != nil {
		regString = *data.Region
	}

	var region models.Region
	if err := database.FirstOrCreate(&region, models.Region{Name: regString}).Error; err != nil {
		return fmt.Errorf("could not find/create Region")
	}

	gatefoldTransparent := false
	if data.GatefoldTransparent != nil {
		gatefoldTransparent = *data.GatefoldTransparent
	}

	var dev models.Developer
	database.Where(models.Developer{Name: string(data.Developer)}).Assign(models.Developer{Slug: slug.Make(string(data.Developer))}).FirstOrCreate(&dev)

	var pub models.Publisher
	database.Where(models.Publisher{Name: string(data.Publisher)}).Assign(models.Publisher{Slug: slug.Make(string(data.Publisher))}).FirstOrCreate(&pub)

	var links []models.Link
	for lt, url := range data.Links {
		var ltype models.LinkType
		database.Where(models.LinkType{SmallName: lt}).Assign(models.LinkType{Name: lt}).FirstOrCreate(&ltype)

		links = append(links, models.Link{
			TypeID: ltype.ID,
			Link:  url,
		})
	}

	game := models.Game{
		Title:       data.Title,
		Slug:        slugTitle,
		Description: data.Description,
		PlatformID:  platform.ID,
		MobygamesID: data.MobygamesId,
		IgdbID: data.IGDBId,
		Links: links,
	}

	database.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"title",
			"description",
			"mobygames_id",
			"igdb_id",
		}),
	}).Create(&game)

	if game.ID == 0 {
		database.Where("slug = ?", slugTitle).First(&game)
	}

	variant := models.Variant{
		GameID:      game.ID,
		Year:        data.Year,
		BoxTypeID:   data.BoxType,
		Description: variantDesc,
		GatefoldTransparent: gatefoldTransparent,
		Slug:        slug.Make(fmt.Sprintf("%s-%s-%d", slugTitle, variantDesc, data.BoxType)), // do I need this?
		DeveloperID: dev.ID,
		PublisherID: pub.ID,
		RegionID:    region.ID,
		Width:       data.Width,
		Height:      data.Height,
		Depth:       data.Depth,
		UserID:      user.ID,
	}

	database.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"year", "description","width","height","depth","gatefold_transparent"}),
	}).Create(&variant)

	variantID := variant.ID
	if variantID == 0 {
		var variant models.Variant
		database.Where("slug = ?", variant.Slug).First(&variant)
		variantID = variant.ID
	}

	docs := []map[string]interface{}{
		{
			"id":   game.ID,
			"slug": game.Slug,
			"variant_id": variantID,
			"title": game.Title,
			"year": variant.Year,
			"region":region.Name,
		},
	}
	pk := "variant_id"
	meili.Index("items").AddDocuments(docs, &meilisearch.DocumentOptions{
		PrimaryKey: &pk,
	})

	// Process images
	var texPaths []string
	wd, err := os.Getwd()
	gameDir := filepath.Join(wd, "uploads/scans", slugTitle, strconv.Itoa(int(variantID)))
	os.MkdirAll(gameDir, os.ModePerm)

	files, err := source.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	foundBox := false
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

		// If we already have glb files, use em!
		if filename == "box.glb" || filename == "box-low.glb" {
			foundBox = true
			tools.Copy(srcPath,gameDir + "/" + filename)
		}
	}

	if !foundBox {

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
	}

	if err := tools.OptimizeWebPImages(texPaths, data.Width, data.Height); err != nil{
		fmt.Errorf("Could not optimize image folder")
	}

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