package main

import (
	"unicode"
	"strings"
	"os"
	"fmt"
	"net/http"
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	layoutISO = "2006-01-02"
)

var (
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbName = os.Getenv("DB_NAME")
	dbUser = os.Getenv("DB_USER")
	dbPass = os.Getenv("DB_PASS")
)

type User struct {
	Name         string `json:"-"`
	DateOfBirth  string `json:"DateOfBirth"`
}

type Birthdays struct {
	gorm.Model
	Name         string `gorm:"column:name"`
	DateOfBirth  string `gorm:"column:bday"`
}

func dbConfig() map[string]string {
	conf := make(map[string]string)
	host, ok := os.LookupEnv("DB_HOST")
	if !ok {
		panic("DB_HOST environment variable required but not set")
	}
	port, ok := os.LookupEnv("DB_PORT")
	if !ok {
		panic("DB_PORT environment variable required but not set")
	}
	user, ok := os.LookupEnv("DB_USER")
	if !ok {
		panic("DB_USER environment variable required but not set")
	}
	password, ok := os.LookupEnv("DB_PASS")
	if !ok {
		panic("DB_PASS environment variable required but not set")
	}
	name, ok := os.LookupEnv("DB_NAME")
	if !ok {
		panic("DB_NAME environment variable required but not set")
	}
	conf[dbHost] = host
	conf[dbPort] = port
	conf[dbUser] = user
	conf[dbPass] = password
	conf[dbName] = name
	return conf
}

func initDb() (db *gorm.DB, err error){
	config := dbConfig()
	connect := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config[dbHost], config[dbPort],
		config[dbUser], config[dbPass], config[dbName])

	db, err = gorm.Open("postgres", connect)
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")
	
	return
}

func onlyLetters(u string) bool {
	for _, s := range u {
		if !unicode.IsLetter(s) {
			return false
		}
	}
	return true
}

func checkDate(date string) bool {
	c := time.Now()
	t, err := time.Parse(layoutISO, date)
	if err != nil {
		fmt.Println(err)
		return false
	}

	// Check the date before current date
	if c.Before(t) {
		log.Warn("The birthday date is incorect",)
		return false
	}
	return true
}

func getDiff(date string) string {
	// Get diff up to bday
	s := ""
	t, _ := time.Parse(layoutISO, date)
	c := time.Now()
	nextYear := time.Date(c.Year() +1, time.December, 31, 0, 0, 0, 0, time.Local)
	days := nextYear.YearDay()
	diff := t.YearDay() - c.YearDay()

	if diff < 0 {
		s = fmt.Sprintf("Your birthday is in %v day(s)!", diff + days)
	} else if diff == 0 {
		s = "Happy birthday!"
	} else {
		s = fmt.Sprintf("Your birthday is in %v day(s)!", diff + days)
	}
	return s
}

func putData(c echo.Context) error {
	user := User{c.Param("username"), ""}

	if !onlyLetters(user.Name) {
		log.Warn("The username must contain only letters"
		return c.String(http.StatusInternalServerError, fmt.Sprintf("The username: %v must contain only letters", user.Name))
	}

	err := c.Bind(&user)
	if err != nil {
		log.Warn("Failed to decode json")
		return c.String(http.StatusInternalServerError, "")
	}

	valid := checkDate(user.DateOfBirth)
	if !valid {
		log.Warn("The date is incorect")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("The date %v is incorect", user.DateOfBirth))
	}

	db, err := initDb()

	dbPut(user, db)
	
	return c.String(http.StatusNoContent, "")
}

func getData(c echo.Context) error {
	user := c.Param("username")

	if !onlyLetters(user) {
		log.Warn("The username: must contain only letters")
		return c.String(http.StatusInternalServerError, fmt.Sprintf("The username: %v must contain only letters", user))
	}

	db, err := initDb()
	if err != nil {
		log.Error("Failed connect to DB")
		return c.String(http.StatusInternalServerError, "Failed connect to DB")
	}
	
	log.Printf(user)
	bday, err := dbGet(user, db)
	if err != nil {
		log.Warn("Failed to get Birthday: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to get Birthday")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": getDiff(bday)})
}

func dbPut(u User, db *gorm.DB) {
	db.AutoMigrate(&Birthdays{})
	user := &Birthdays{Name: strings.ToLower(u.Name), DateOfBirth: u.DateOfBirth}
	db.Where(Birthdays{Name: strings.ToLower(u.Name)}).Assign(Birthdays{Name: strings.ToLower(u.Name), DateOfBirth: u.DateOfBirth}).FirstOrCreate(&user)
}

func dbGet(username string, db *gorm.DB) (bday string, err error) {
	log.Infof(username)
	defer db.Close()
	bday = ""
	row := db.Table("birthdays").Where("name = ?", strings.ToLower(username)).Select("name, bday").Row()
	err = row.Scan(&username, &bday)
	if err != nil {
		return bday, err
	}
	return bday, nil
}

func main() {

	Formatter := new(log.JSONFormatter)
	log.SetFormatter(Formatter)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
        
	e.GET("/hello/:username", getData)
	e.PUT("/hello/:username", putData)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.Logger.Fatal(e.Start(":8000"))
}
