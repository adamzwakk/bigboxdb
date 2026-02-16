package tools

import (
	"math/rand"
	"io"
	"os"
	"fmt"
	"encoding/json"
)

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
			return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
			return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
			return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
			return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

type ImportData struct{
	Title			string	`json:"title"`
	Description		*string	`json:"description,omitempty"`
	Region			*string `json:"region,omitempty"`
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
	GatefoldTransparent		*bool `json:"gatefold_transparent"`
	IGDBId			*int		`json:"igdb_id,omitempty"`
	IgdbSlug		*string		`json:"igdb_slug,omitempty"`
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