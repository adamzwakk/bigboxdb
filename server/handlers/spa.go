package handlers

import (
    "fmt"
    "os"
    "strings"

    "github.com/gin-gonic/gin"
)

// This mostly exists for the docker container hosting reasons but you could use it I guess
func ServeIndex(c *gin.Context) {
    path := c.Request.URL.Path

    if strings.HasPrefix(path, "/game/") {
        slug := strings.TrimPrefix(path, "/game/")
        if m, ok := GetMeta(slug); ok {
            serveWithMeta(c, m)
            return
        }
    }

    c.File("./web/index.html")
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