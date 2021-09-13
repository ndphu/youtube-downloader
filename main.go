package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
	"youtube-downloader/download"
)

func main() {
	r := gin.Default()

	c := cors.DefaultConfig()
	c.AllowAllOrigins = true
	c.AllowCredentials = true
	c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	c.AllowHeaders = []string{"Origin", "Authorization", "Content-Type", "Content-Length", "X-Requested-With"}

	r.Use(cors.New(c))
	api := r.Group("/api")
	{
		api.GET("/video/:key", func(c *gin.Context) {
			id := c.Param("key")
			log.Println("Crawling video", id)
			v := download.NewVideo(id)
			if err := v.Load(); err != nil {
				panic(err)
			}
			log.Println(v.Info.Title)
			stream, err := v.GetAudioStream()
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"success": false, "error": err.Error()})
			} else {
				log.Println("Audio", stream)
				c.JSON(200, gin.H{"success": true, "title": v.Info.Title, "audio": stream})
			}
		})
	}

	r.Run()
}
