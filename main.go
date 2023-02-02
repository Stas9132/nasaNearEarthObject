package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const url = ":3000"

type service struct {
	store map[string]string
	db    *gorm.DB
}

var c = make(chan string, 10*10)

func (s *service) getHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	storeUpdateFlag := false
	for _, date := range r.Form["dates"] {
		rx := regexp.MustCompile("(?:20[0-4][0-9])-(?:1[0-2]|0[1-9])")
		month := rx.FindString(date)
		if len(month) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, ok := s.store[date]
		if !ok {
			c <- date
			storeUpdateFlag = true
		}
	}

	if storeUpdateFlag {
		time.Sleep(time.Second)
	}

	var response []string
	w.WriteHeader(http.StatusOK)
	for _, date := range r.Form["dates"] {
		rx := regexp.MustCompile("(?:20[0-4][0-9])-(?:1[0-2]|0[1-9])")
		month := rx.FindString(date)
		e, ok := s.store[month]
		response = append(response, e)
		if !ok {
			w.WriteHeader(http.StatusPartialContent)
		}
	}
	w.Write([]byte(strings.Join(response, " ")))
	return
}

func (s *service) storeUpdateFunc() {
	for reqMonth := range c {
		_, ok := s.store[reqMonth]
		if !ok {
			startTime, _ := time.Parse("2006-01-02", reqMonth+"-01")
			startTime.da
			endTime := startTime.AddDate(0, 1, -1)

			startDate := startTime.Format("2006-01-02")
			endDate := endTime.Format("2006-01-02")
			s.store[reqMonth] = startDate + "w" + endDate
		}
	}
}

type NeoCountsTable struct {
	Records []NeoCount `json:"neo_counts"`
}

type NeoCount struct {
	Date  string `json:"date" gorm:"primary_key"`
	Count int    `json:"count"`
}

func (s *service) postHandler(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer r.Body.Close()

	var t NeoCountsTable
	if err = json.Unmarshal(content, &t); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	for _, neoCountRecord := range t.Records {
		s.db.Create(&neoCountRecord)
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("ok"))
	return
}

func main() {
	fmt.Println("Сервис выводящий общее количество астероидов в околоземном пространстве на определённую дату")

	srv := service{store: make(map[string]string)}
	var err error
	srv.db, err = gorm.Open("mysql", "user:12345678@tcp(127.0.0.1:3306)/db1?charset=utf8&parseTime=True")
	if err != nil {
		log.Fatal(err)
	}
	srv.db.AutoMigrate(&NeoCount{})

	go srv.storeUpdateFunc()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/neo/count", srv.getHandler)
	r.Post("/neo/count", srv.postHandler)

	if err := http.ListenAndServe(url, r); err != nil {
		log.Fatal(err)
	}
}
