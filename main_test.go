package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCafeNegative(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []struct {
		request string
		status  int
		message string
	}{
		{"/cafe", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=omsk", http.StatusBadRequest, "unknown city"},
		{"/cafe?city=tula&count=na", http.StatusBadRequest, "incorrect count"},
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v.request, nil)
		handler.ServeHTTP(response, req)

		assert.Equal(t, v.status, response.Code)
		assert.Equal(t, v.message, strings.TrimSpace(response.Body.String()))
	}
}

func TestCafeWhenOk(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)

	requests := []string{
		"/cafe?count=2&city=moscow",
		"/cafe?city=tula",
		"/cafe?city=moscow&search=ложка",
	}
	for _, v := range requests {
		response := httptest.NewRecorder()
		req := httptest.NewRequest("GET", v, nil)

		handler.ServeHTTP(response, req)

		assert.Equal(t, http.StatusOK, response.Code)
	}
}

// TestCafeCount validates the pagination logic of the cafe handler.
// It checks boundary conditions for the "count" parameter, including
// zero values, small increments, and values exceeding the total
// number of cafes available in the data set.
func TestCafeCount(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	totalCount := len(cafeList[city])

	requests := []struct {
		name  string
		count int
		want  int
	}{
		{
			name:  "Count is 0",
			count: 0,
			want:  0,
		},
		{
			name:  "Count is 1",
			count: 1,
			want:  1,
		},
		{
			name:  "Count is 2",
			count: 2,
			want:  2,
		},
		{
			name:  "Count is 100 (more than total)",
			count: 100,
			want:  totalCount,
		},
	}

	for _, v := range requests {
		t.Run(v.name, func(t *testing.T) {
			params := url.Values{}
			params.Add("city", city)
			params.Add("count", strconv.Itoa(v.count))

			path := "/cafe?" + params.Encode()

			req := httptest.NewRequest("GET", path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()

			var actualCount int
			if body == "" {
				actualCount = 0
			} else {
				cafes := strings.Split(body, ",")
				actualCount = len(cafes)
			}

			assert.Equal(t, v.want, actualCount, "For count=%d expected %d cafes, got %d", v.count, v.want, actualCount)
		})
	}
}

// TestCafeSearch verifies the search functionality for cafes within a specific city.
// It uses a table-driven approach to validate:
// 1. Empty results for non-matching queries.
// 2. Case-insensitive partial string matching.
// 3. Presence of the search term in every returned cafe name.
func TestCafeSearch(t *testing.T) {
	handler := http.HandlerFunc(mainHandle)
	city := "moscow"

	requests := []struct {
		name      string
		search    string
		wantCount int
	}{
		{
			name:      "Search 'фасоль' - no results",
			search:    "фасоль",
			wantCount: 0,
		},
		{
			name:      "Search 'кофе' - two results",
			search:    "кофе",
			wantCount: 2,
		},
		{
			name:      "Search 'вилка' - one result",
			search:    "вилка",
			wantCount: 1,
		},
	}

	for _, v := range requests {
		t.Run(v.name, func(t *testing.T) {
			params := url.Values{}
			params.Add("city", city)
			params.Add("search", v.search)

			path := "/cafe?" + params.Encode()

			req := httptest.NewRequest("GET", path, nil)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, req)

			assert.Equal(t, http.StatusOK, response.Code)

			body := response.Body.String()
			var cafes []string
			if body != "" {
				cafes = strings.Split(body, ",")
			}

			assert.Len(t, cafes, v.wantCount, "Search for '%s' expected %d cafes, got %d", v.search, v.wantCount, len(cafes))

			for _, cafeName := range cafes {
				assert.Containsf(t, strings.ToLower(cafeName), strings.ToLower(v.search), "Cafe name '%s' must contain search string '%s'", cafeName, v.search)
			}
		})
	}
}
