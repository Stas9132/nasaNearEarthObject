package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetHandler(t *testing.T) {
	srv := service{store: make(map[string]int)}
	srv.store["2020-08-22"] = 10
	srv.store["2020-03-20"] = 20
	srv.store["2020-05-24"] = 30
	result := "60"

	req := httptest.NewRequest(http.MethodGet, "/neo/count", nil)
	req.Form = url.Values{}
	req.Form.Add("dates", "2020-08-22")
	req.Form.Add("dates", "2020-03-20")
	req.Form.Add("dates", "2020-05-24")
	w := httptest.NewRecorder()

	srv.getHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	if result != string(body) {
		t.Log("Expired result", result, "received", string(body))
		t.Fail()
	}
}

func TestPostHandler(t *testing.T) {
	srv := service{store: make(map[string]int)}
	jsonReq := "{\"neo_counts:\":[{\"date\":\"2020-01-20\",\"count\":12},{\"date\":\"2020-02-26\",\"count\":9}]}"
	result := "ok"

	req := httptest.NewRequest(http.MethodPost, "/neo/count", strings.NewReader(jsonReq))
	w := httptest.NewRecorder()

	srv.postHandler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	fmt.Println(resp.Status)
	defer resp.Body.Close()

	if result != string(body) {
		t.Log("Expired result", result, "received", string(body))
		t.Fail()
	}
}
