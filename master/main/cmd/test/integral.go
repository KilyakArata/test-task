package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type Banner struct {
	BannerId  int               `json:"banner_id,omitempty"`
	TagIds    []int             `json:"tag_ids,omitempty"`
	FeatureId int               `json:"feature_id,omitempty"`
	Content   map[string]string `json:"content,omitempty"`
	IsActive  bool              `json:"is_active,omitempty"`
	CreatedAt string            `json:"created_at,omitempty"`
	UpdatedAt string            `json:"updated_at,omitempty"`
}

func main() {
	tests := []struct {
		name           string
		nameFunc       string
		token          string
		method         string
		url            string
		numberOfBanner int
		body           interface{}
		banner         map[string]string
		expectedStatus int
	}{
		{
			name:     "PostBanner101 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{101, 102, 103},
				FeatureId: 101,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 101",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner102 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{103, 104},
				FeatureId: 102,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 102",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner103 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{105, 106},
				FeatureId: 103,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 103",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner104 not active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{101, 102, 103, 104},
				FeatureId: 104,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 104",
					"url":   "https://example.com/article",
				},
				IsActive: false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner105 not active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{103, 104},
				FeatureId: 105,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 105",
					"url":   "https://example.com/article",
				},
				IsActive: false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner106 not active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{106, 107},
				FeatureId: 101,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 106",
					"url":   "https://example.com/article",
				},
				IsActive: false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner107 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{106, 107},
				FeatureId: 104,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 107",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner108 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{101, 102},
				FeatureId: 105,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 108",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner109 not active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{106, 107},
				FeatureId: 102,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 109",
					"url":   "https://example.com/article",
				},
				IsActive: false,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "PostBanner110 active",
			nameFunc: "PostBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "POST",
			url:      "http://localhost:8181/banner",
			body: Banner{
				TagIds:    []int{105},
				FeatureId: 101,
				Content: map[string]string{
					"text":  "Текст статьи",
					"title": "Заголовок 110",
					"url":   "https://example.com/article",
				},
				IsActive: true,
			},
			expectedStatus: http.StatusOK,
		},

		{
			name:     "GetBanner last version",
			nameFunc: "GetBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "GET",
			url:      "http://localhost:8181/user_banner?tag_id=1&feature_id=1&use_last_revision=true",
			body:     Banner{},
			banner: map[string]string{
				"text":  "Текст статьи",
				"title": "Заголовок 1",
				"url":   "https://example.com/article"},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "GetBanner any version",
			nameFunc: "GetBanner",
			token:    "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:   "GET",
			url:      "http://localhost:8181/user_banner?tag_id=1&feature_id=1",
			body:     Banner{},
			banner: map[string]string{
				"text":  "Текст статьи",
				"title": "Заголовок 1",
				"url":   "https://example.com/article"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "DeleteBanner101",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner102",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner103",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner104",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner105",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner106",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner107",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner108",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner109",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "DeleteBanner110",
			nameFunc:       "DeleteBanner",
			token:          "c1c224b03cd9bc7b6a86d77f5dace40191766c485cd55dc48caf9ac873335d6f",
			method:         "DELETE",
			url:            "http://localhost:8181/banner/",
			expectedStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		count := 0
		if tt.nameFunc == "DeleteBanner" {
			tt.url += strconv.Itoa(tests[count].numberOfBanner)
			count++
		}
		req, err := http.NewRequest(tt.method, tt.url, nil)
		if err != nil {
			fmt.Println("ошибка http.NewRequest: ", err)
			return
		}
		req.Header.Add("token", tt.token)
		reqBody, err := json.Marshal(tt.body)
		if err != nil {
			fmt.Println("ошибка json.Marshal: ", err)
			return
		}
		req.Body = io.NopCloser(bytes.NewReader(reqBody))

		client := http.DefaultClient

		result, err := client.Do(req)
		if err != nil {
			fmt.Println("ошибка client.Do: ", err)
			return
		}

		bodyFromServer, err := io.ReadAll(result.Body)
		if err != nil {
			fmt.Println("ошибка io.ReadAll: ", err)
			return
		}
		if tt.nameFunc == "PostBanner" {
			var id int
			err = json.Unmarshal(bodyFromServer, &id)
			if err != nil {
				fmt.Println("ошибка json.Unmarshal: ", err)
				return
			}
			tt.numberOfBanner = id
		}
		if tt.nameFunc == "GetBanner" {
			var banner map[string]string
			err = json.Unmarshal(bodyFromServer, &banner)
			if err != nil {
				fmt.Println("ошибка json.Unmarshal: ", err)
				return
			}
			for k := range tt.banner {
				if _, ok := banner[k]; !ok {
					fmt.Printf("%v - нет такого ключа\n", k)
					return
				}
				if tt.banner[k] != banner[k] {
					fmt.Printf("%v - не равно - %v\n", tt.banner[k], banner[k])
					return
				}
			}
		}
		status := result.StatusCode
		if status != tt.expectedStatus {
			fmt.Printf("%v - не равно - %v\n", tt.expectedStatus, status)
			return
		}
		fmt.Printf("тест: %v, - прошёл успешно\n", tt.name)
		result.Body.Close()
	}
	fmt.Println("интеграционный тест выполнился успешно")
}
