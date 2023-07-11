package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

// Config struct to hold the configuration values
type Config struct {
	Database struct {
		Password string `yaml:"password"`
		User     string `yaml:"user"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		DBName   string `yaml:"dbName"`
	} `yaml:"database"`
}

// Request input params
type Request struct {
	Login string `json:"login"`
}

// Response output
type Response struct {
	Message string `json:"message"`
}

type StudentCompletion struct {
	StuID                string  `json:"stu_id"`
	StuName              string  `json:"stu_name"`
	TotalCourses         int     `json:"total_courses"`
	CoursesCompleted     int     `json:"courses_completed"`
	CompletionPercentage float64 `json:"completion_percentage"`
}

func Validate(login string) bool {
	if len(login) < 2 || len(login) > 10 {
		return false
	}

	for _, c := range login {
		if !unicode.IsLower(c) && !unicode.IsNumber(c) {
			return false
		}
	}

	return true
}

func setupDB(config Config) (*sql.DB, error) {
	user := config.Database.User
	host := config.Database.Host
	port := config.Database.Port
	dbName := config.Database.DBName
	password := config.Database.Password

	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, dbName)

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func calculateCompletionPercentage(totalCourses, coursesCompleted int) float64 {
	if totalCourses == 0 {
		return 0.0
	}

	return float64(coursesCompleted) / float64(totalCourses) * 100
}

func main() {
	configFile := flag.String("config", "config.yaml", "Path")
	flag.Parse()

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	for {
		db, err := setupDB(config)
		if err != nil {
			fmt.Println("Erreur lors de la tentative de connexion à la base de données:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer db.Close()

		gin.SetMode(gin.ReleaseMode)
		r := gin.Default()
		r.POST("/webservice", func(c *gin.Context) {
			// Vérifier l'en-tête X-API-KEY
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

			if !Validate(req.Login) {
				c.JSON(400, gin.H{"error": "Invalid login"})
				return
			}

			res, err := db.Query("SELECT s.stu_id, s.stu_name, COUNT(DISTINCT m.course_id) AS total_courses, COUNT(DISTINCT m.course_id) AS courses_completed FROM student s LEFT JOIN module m ON s.stu_id = m.stu_id WHERE s.stu_id = ? GROUP BY s.stu_id, s.stu_name;", req.Login)
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to execute database query"})
				return
			}
			defer res.Close()

			var studentCompletion StudentCompletion
			if res.Next() {
				err := res.Scan(
					&studentCompletion.StuID,
					&studentCompletion.StuName,
					&studentCompletion.TotalCourses,
					&studentCompletion.CoursesCompleted,
				)
				if err != nil {
					c.JSON(500, gin.H{"error": "Failed to scan database result"})
					return
				}

				studentCompletion.CompletionPercentage = calculateCompletionPercentage(studentCompletion.TotalCourses, studentCompletion.CoursesCompleted)
			} else {
				c.JSON(404, gin.H{"error": "Student not found"})
				return
			}

			c.JSON(200, studentCompletion)
		})

		fmt.Println("Waiting for requests....")
		r.Run(":8000")
	}
}
