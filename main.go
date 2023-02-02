package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/peteretelej/nasa"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const url = ":3000"

type service struct {
	store map[string]int
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
		_, ok := s.store[date]
		if !ok {
			c <- date
			storeUpdateFlag = true
		}
	}

	if storeUpdateFlag {
		time.Sleep(10 * time.Second)
	}

	var response int
	w.WriteHeader(http.StatusOK)
	for _, date := range r.Form["dates"] {
		e, ok := s.store[date]
		if ok {
			response += e
		} else {
			w.WriteHeader(http.StatusPartialContent)
		}
	}
	w.Write([]byte(strconv.Itoa(response)))
	return
}

func (s *service) updateStoreFromNasa() {
	nlc := make(chan *nasa.NeoList)
	go func() {
		for nl := range nlc {
			for date, asteroids := range nl.NearEarthObjects {
				s.store[date] = len(asteroids)
			}
		}
	}()
	for reqDate := range c {
		_, ok := s.store[reqDate]
		if !ok {
			startTime, err := time.Parse("2006-01-02", reqDate)
			if err != nil {
				continue
			}
			endTime := startTime.AddDate(0, 0, 0)

			go func(sT, eT time.Time) {
				neoList, err := nasa.NeoFeed(startTime, endTime)
				if err == nil {
					nlc <- neoList
				}
			}(startTime, endTime)
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

var sc = make(chan NeoCount, 10)

func (s *service) postHandler(w http.ResponseWriter, r *http.Request) {
	content, err := io.ReadAll(r.Body)
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
		sc <- neoCountRecord
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("ok"))
	return
}

func statsUpload() {
	dsn := "user:12345678@tcp(127.0.0.1:3306)/db1?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = db.AutoMigrate(&NeoCount{})
	if err != nil {
		panic(err)
	}
	for neoCountRecord := range sc {
		db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&neoCountRecord)
	}
}

func main() {
	fmt.Println("Сервис выводящий общее количество астероидов в околоземном пространстве на определённую дату")

	srv := service{store: make(map[string]int)}

	go srv.updateStoreFromNasa()
	go statsUpload()

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/neo/count", srv.getHandler)
	r.Post("/neo/count", srv.postHandler)

	if err := http.ListenAndServe(url, r); err != nil {
		log.Fatal(err)
	}
}
