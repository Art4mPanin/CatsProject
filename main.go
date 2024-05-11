package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Cat struct {
	Id     string `json:"id"`
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type CatReq struct {
	Name     string `json:"name"`
	Age      int    `json:"age"`
	Homeless bool   `json:"homeless"`
	ImgURL   string `json:"img_url"`
}
type CatResponse struct {
	CatReq
	ID int `json:"id"`
}

var globalURL = ""

func updateURL() {
	for {
		res, err := http.Get("https://api.thecatapi.com/v1/images/search")
		if err != nil {
			log.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
		}
		if err != nil {
			log.Fatal(err)
		}
		var cats []Cat
		err = json.Unmarshal(body, &cats)
		if err != nil {
			fmt.Println("error:", err)
		}
		err = res.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		if err != nil {
			log.Fatal(err)
		}
		globalURL = cats[0].Url
		fmt.Println("globalURL updated")
		time.Sleep(60 * time.Second)
	}
}
func createTablesIfNotExists() {
	database, err := sql.Open("sqlite3", "godb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS Feedback(id INTEGER PRIMARY KEY AUTOINCREMENT, url TEXT, quality INT, cute INT, message TEXT)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}

func FeedbackHandler(c echo.Context) error {
	name := c.QueryParam("quality")
	name1 := c.QueryParam("cute")
	name2 := c.QueryParam("message")
	database, err := sql.Open("sqlite3", "godb.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	q, err := strconv.Atoi(name)
	cu, err := strconv.Atoi(name1)
	if err != nil {
		return c.HTML(http.StatusBadRequest, fmt.Sprintf("<p>quality and cute must be numbers</p>"))
	}

	statement, err := database.Prepare("INSERT INTO Feedback(url, quality, cute, message) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec(globalURL, q, cu, name2)
	if err != nil {
		log.Fatal(err)
	}
	return c.String(http.StatusOK, fmt.Sprintf("Quality - %v\nCute - %v\nMessage - %v", name, name1, name2))
}
func CatHandler(c echo.Context) error {
	return c.HTML(http.StatusOK, fmt.Sprintf("<h1>Это твой рандомный кот в эту минуту</h1><img src=%v>", globalURL))
}

func addCat(c echo.Context) error {
	cat := CatReq{}
	err := c.Bind(&cat)
	if err != nil {
		log.Printf("Failed processing addCat request: %s", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	log.Printf("This is ur cat: %v", cat)
	go AddingFromPost(cat)
	return c.String(http.StatusOK, fmt.Sprintf("data was added"))
}
func createTablesIfNotExists1() {
	database, err := sql.Open("sqlite3", "Catinfo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()
	statement, err := database.Prepare("CREATE TABLE IF NOT EXISTS Catinfo(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, age INT, homeless BOOL, img_url TEXT, created_at TIMESTAMP)")
	if err != nil {
		log.Fatal(err)
	}
	_, err = statement.Exec()
	if err != nil {
		log.Fatal(err)
	}
}
func AddingFromPost(cat CatReq) {
	name := cat.Name
	age := cat.Age
	home := cat.Homeless
	//img_url := cat.ImgURL
	database, err := sql.Open("sqlite3", "Catinfo.db")
	if err != nil {
		log.Fatalf("There is an error occured in opening db: %s", err)
	}
	defer database.Close()
	statement, err := database.Prepare("INSERT INTO Catinfo(name, age, homeless, img_url, created_at) VALUES (?,?,?,?, current_time)")
	if err != nil {
		log.Fatalf("There is an error occured in preparing statement: %s", err)
	}
	_, err = statement.Exec(name, age, home, globalURL)
	if err != nil {
		log.Fatal(err)
	}
}
func GetData(c echo.Context) error {
	db, err := sql.Open("sqlite3", "Catinfo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlQuery := "SELECT id, name, age, homeless, img_url FROM Catinfo ORDER BY id DESC "
	limit := c.QueryParam("limit")
	if limit != "" {
		if _, err := strconv.Atoi(limit); err == nil {
			sqlQuery += "LIMIT " + limit
		}
	}
	rows, err := db.Query(sqlQuery)
	if err != nil {
		log.Panic(err)
	}
	defer rows.Close()
	var Infos []CatResponse
	for rows.Next() {
		cat := CatResponse{}
		err = rows.Scan(&cat.ID, &cat.Name, &cat.Age, &cat.Homeless, &cat.ImgURL)
		if err != nil {
			fmt.Println(err)
			continue
		}
		Infos = append(Infos, cat)
	}
	return c.JSON(http.StatusOK, Infos)
}

func IdGetData(c echo.Context) error {
	db, err := sql.Open("sqlite3", "Catinfo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	sqlQuery := "SELECT DISTINCT id, name, age, homeless, img_url FROM Catinfo"
	id := c.Param("id")
	if id != "" {
		if _, err := strconv.Atoi(id); err == nil {
			sqlQuery += " WHERE id =" + id
		}
	} else {
		return c.String(http.StatusBadRequest, fmt.Sprintf("U gave no id"))
	}
	row := db.QueryRow(sqlQuery)
	cat := CatResponse{}
	err = row.Scan(&cat.ID, &cat.Name, &cat.Age, &cat.Homeless, &cat.ImgURL)
	if errors.Is(err, sql.ErrNoRows) {

		return c.String(http.StatusNotFound, fmt.Sprintf("Такого кота нету, вы - еблан"))
	}
	return c.HTML(http.StatusOK, fmt.Sprintf("<h1>Это кот - %s</h1><h1>Это данные: %v\n,%v\n,%v\n,%v\n,%v\n</h1><img src=%v>", cat.Name, cat.ID, cat.Name, cat.Age, cat.Homeless, cat.ImgURL, cat.ImgURL))
}

func main() {
	//get url
	go updateURL()
	createTablesIfNotExists()
	createTablesIfNotExists1()
	//server
	e := echo.New()
	e.GET("/api/cats", CatHandler)
	e.GET("/api/cats/feedback", FeedbackHandler)

	e.POST("/api/cats/internal/create", addCat)
	e.GET("/api/cats/internal/list", GetData)
	e.GET("/api/cats/internal/:id/view", IdGetData)
	e.Logger.Fatal(e.Start(":1324"))
}
