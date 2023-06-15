package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

// Config struct to hold the configuration values
type Config struct {
	Database struct {
		Password string `yaml:"password"`
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

func setupDB(password string) (*sql.DB, error) {
	user := "root"
	host := "localhost"
	port := "3306"
	dbName := "test"

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

	password := config.Database.Password

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-config") {
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	db, err := setupDB(password)
	if err != nil {
		fmt.Println("Erreur lors de la tentative de connexion à la base de données:", err)
		return
	}
	defer db.Close()

	res, err := db.Query("SELECT s.stu_id, s.stu_name, COUNT(DISTINCT m.course_id) AS total_courses, COUNT(DISTINCT m.course_id) AS courses_completed, ROUND((COUNT(DISTINCT m.course_id) / NULLIF(COUNT(DISTINCT m.course_id), 0)) * 100, 2) AS completion_percentage FROM student s LEFT JOIN module m ON s.stu_id = m.stu_id GROUP BY s.stu_id, s.stu_name;")

	if err != nil {
		fmt.Println("Erreur lors de l'exécution de la requête:", err)
		return
	}
	defer res.Close()

	var studentCompletionList []StudentCompletion

	for res.Next() {
		var studentCompletion StudentCompletion
		err := res.Scan(
			&studentCompletion.StuID,
			&studentCompletion.StuName,
			&studentCompletion.TotalCourses,
			&studentCompletion.CoursesCompleted,
			&studentCompletion.CompletionPercentage,
		)

		if err != nil {
			fmt.Println("Colonne non trouvée", err)
			continue
		}
		studentCompletionList = append(studentCompletionList, studentCompletion)
	}

	if err := res.Err(); err != nil {
		fmt.Println("Une erreur s'est produite:", err)
		return
	}

	for _, sc := range studentCompletionList {
		fmt.Printf("ID étudiant: %s\n", sc.StuID)
		fmt.Printf("Nom étudiant: %s\n", sc.StuName)
		fmt.Printf("Cours totaux: %d\n", sc.TotalCourses)
		fmt.Printf("Cours complétés: %d\n", sc.CoursesCompleted)
		fmt.Printf("Pourcentage de cours complétés: %.2f%%\n", sc.CompletionPercentage)
		fmt.Println("---------------")
	}

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

		out, err := exec.Command("echo", "zmprov", "ma", req.Login, "1").Output()
		if err != nil {
			c.String(500, err.Error())
			return
		}
		resp := Response{Message: string(out)}
		c.JSON(200, resp)
	})

	fmt.Println("Waiting for requests....")
	r.Run(":8000")
}
