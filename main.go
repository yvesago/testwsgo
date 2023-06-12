package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os/exec"
)

// Request  input params
type Request struct {
	Login string `json:"login"`
}

// Response output
type Response struct {
	Message string `json:"message"`
}

func Validate(string) bool {
	 return false
}

func main() {
	r := gin.Default()
	r.POST("/webservice", func(c *gin.Context) {
		apiKeyReceived := c.GetHeader("X-API-KEY")
		if apiKeyReceived != "mysecretkey" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		var req Request
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		out, err := exec.Command("echo", "zmprov", "ma", req.Login, "1").Output()
		if err != nil {
			c.String(500, err.Error())
			return
		}
		resp := Response{Message: string(out)}
		c.JSON(200, resp)
	})
	fmt.Println("waiting for requests....")
	r.Run(":8000")
}
