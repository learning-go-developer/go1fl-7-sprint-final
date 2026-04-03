package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
			url := fmt.Sprintf("/cafe?city=%s&count=%d", city, v.count)

			req := httptest.NewRequest("GET", url, nil)
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
