package main

import (
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // Enable CORS for frontend
    r.Use(cors.Default())

    r.GET("/api/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "Hello from Gin!",
        })
    })

    r.Run(":8080")
}