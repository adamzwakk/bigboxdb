package handlers

import (
    "fmt"
    "os"
    "strings"
    "strconv"
    "log"

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
    html, err := os.ReadFile("./web/index.html")
    if err != nil {
        c.File("./web/index.html")
        return
    }

    tags := fmt.Sprintf(`<title>%s</title>
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:image" content="%s">
</head>`, m.Title, m.Title, m.Description, m.Image)

    modified := strings.Replace(string(html), "</head>", tags, 1)
    c.Header("Content-Type", "text/html")
    c.String(200, modified)
}