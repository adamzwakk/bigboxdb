package db

import (
	"os"
	"log"
	"fmt"

    "github.com/meilisearch/meilisearch-go"
)

func InitMeiliSearch() meilisearch.ServiceManager {
    return meilisearch.New(os.Getenv("MEILI_URL"), meilisearch.WithAPIKey(os.Getenv("MEILI_MASTER_KEY")))
}

func InitMeilisearchPublic() {
	client := InitMeiliSearch()
    // Check if key exists, if not create it
    keys, _ := client.GetKeys(nil)
    
    for _, key := range keys.Results {
        if key.Description == "Search-only key" {
			fmt.Println(key.Key)
            return
        }
    }
    
    key, err := client.CreateKey(&meilisearch.Key{
		Description: "Search-only key",
		Actions:     []string{"search"},
		Indexes:     []string{"*"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("API Key:", key.Key)
}