package main

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"os"
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

	r.GET("/watch", func(c *gin.Context) {
		id := c.Query("v")
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

		api.GET("/video/:key/next", func(c *gin.Context) {
			id := c.Param("key")
			log.Println("Crawling next video", id)
			v := download.NewVideo(id)
			if v, err := v.Next(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"success": false, "error": err.Error()})
				return
			} else {
				c.JSON(200, gin.H{"success": true, "videoId": v})
			}
		})

		api.GET("/video/:key/nextList", func(c *gin.Context) {
			id := c.Param("key")
			log.Println("Crawling next video", id)
			v := download.NewVideo(id)
			if videos, err := v.NextList(); err != nil {
				c.AbortWithStatusJSON(500, gin.H{"success": false, "error": err.Error()})
				return
			} else {
				c.JSON(200, gin.H{"success": true, "videos": videos})
			}

		})

		api.GET("/list/:key", func(c *gin.Context) {

		})

		api.GET("/cache", func(c *gin.Context) {
			decipher, err := ioutil.ReadFile(".cache.decipherOps")
			if err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Fail to read \".cache.decipherOps\"", "error": err.Error()})
				return
			}
			signature, err := ioutil.ReadFile(".cache.signature")
			if err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Fail to read \".cache.signature\"", "error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"success": true, "decipher": string(decipher), "signature": string(signature)})
		})

		api.GET("/clearCache", func(c *gin.Context) {
			if err := os.Remove(".cache.decipherOps"); err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Fail to remove \".cache.decipherOps\"", "error": err.Error()})
				return
			}
			if err := os.Remove(".cache.signature"); err != nil {
				c.JSON(500, gin.H{"success": false, "message": "Fail to remove \".cache.signature\"", "error": err.Error()})
				return
			}

			c.JSON(200, gin.H{"success": true})
		})
	}

	r.Run()
}
