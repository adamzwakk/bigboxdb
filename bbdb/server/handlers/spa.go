package handlers

import (
    "fmt"
    "os"
    "strings"
    "strconv"
	"regexp"
    // "log"

    "github.com/gin-gonic/gin"
)

func ServeIndex(c *gin.Context) {
    path := c.Request.URL.Path

    if strings.HasPrefix(path, "/game/") {
        remainder := strings.TrimPrefix(path, "/game/")
        parts := strings.Split(remainder, "/")

        slug := parts[0]
        var variantID int = 0
        if len(parts) > 1 {
            variantID, _ = strconv.Atoi(parts[1])
        }
        if m, ok := GetMeta(slug, variantID); ok {
            serveWithMeta(c, m)
            return
        }
    }

    return
}

func serveWithMeta(c *gin.Context, m Meta) {
    indexPath := "./web/index.html"
    if os.Getenv("APP_ENV") == "production" {
        indexPath = "/usr/share/nginx/html/index.html"
    }
    html, err := os.ReadFile(indexPath)
    if err != nil {
        c.File(indexPath)
        return
    }

    re := regexp.MustCompile(`<title>.*?</title>`)
	modified := re.ReplaceAllString(string(html), fmt.Sprintf(`<title>%s | BigBoxDB</title>`, m.Title))

	// Inject OG meta tags before </head>
	ogTags := fmt.Sprintf(`<meta property="og:title" content="%s | BigBoxDB">
		<meta property="og:description" content="%s">
		<meta property="og:image" content="%s">
	</head>`, m.Title, m.Description, m.Image)
	modified = strings.Replace(modified, "</head>", ogTags, 1)
    c.Header("Content-Type", "text/html")
    c.String(200, modified)
}