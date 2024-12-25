package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type coord struct {
	Lon float64 `json:"lon"`
	Lat float64 `json:"lat"`
}

type city struct {
	Id         float64 `json:"id"`
	Name       string  `json:"name"`
	State      string  `json:"state"`
	Country    string  `json:"country"`
	Coord      coord   `json:"coord"`
	Population uint64  `json:"population"`
}
type timezone struct {
	U int      `json:"u"`
	C []string `json:"c"`
	A string   `json:"a"`
	R int      `json:"r"`
	D int      `json:"d"`
}

var (
	cities       []city
	timezones    map[string]timezone
	countryCodes map[string]string
)

func importLocalJson[E any](filename string, v E) E {
	log.Printf("Importing json data: %s %v\n", filename, v)
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("unable to read: %s", filename))
	}

	if err := json.Unmarshal(data, v); err != nil {
		panic(err)
	}

	return v
}

type CountryCode struct {
	Country string `json:"country"`
	Code    string `json:"code"`
}

func main() {
	importLocalJson("city.list.json", &cities)
	importLocalJson("timezone.json", &timezones)
	importLocalJson("country-code.json", &countryCodes)

	fmt.Printf("total cities: %d\n", len(cities))
	fmt.Printf("total timezones: %d\n", len(timezones))
	fmt.Printf("total country codes: %d\n", len(countryCodes))

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		str := r.URL.Query().Get("query")
		if str != "" {
			matched := []city{}
			for i := 0; i < len(cities); i++ {
				city := cities[i]

				if strings.Contains(strings.ToLower(city.Name), strings.ToLower(str)) {
					matched = append(matched, city)
				}
			}

			w.WriteHeader(200)
			if err := json.NewEncoder(w).Encode(matched); err != nil {
				panic(err)
			}

			return
		}

		timezone := r.URL.Query().Get("timezone")
		if timezone != "" {
			t, exists := timezones[timezone]
			if exists {
				code := t.C[0]

				w.WriteHeader(200)
				if err := json.NewEncoder(w).Encode(
					&CountryCode{
						Code:    code,
						Country: countryCodes[code],
					},
				); err != nil {
					panic(err)
				}
			} else {
				w.WriteHeader(404)
				if err := json.NewEncoder(w).Encode(&countryCodes); err != nil {
					panic(err)
				}
			}
		}
	})

	log.Println("Server is running on port 3000")
	http.ListenAndServe(":3000", r)
}
