package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"
)

// TODO: custom expire times, maybe as a flag
// TODO: Single Use flag
// TODO: Minimal UI
// TODO: Throttle

type Lock struct{ sync.Mutex }

func (m *Lock) Within(f func()) {
	m.Lock()
	defer m.Unlock()
	f()
}

var (
	port = flag.String("port", "4242", "listen port")

	pastes        = map[string]string{}
	expireTime    = 2 * time.Minute
	expireTracker = map[string]time.Time{}
	lock          Lock
)

func main() {
    flag.Parse()

	go func() {
		for {
			lock.Within(func() {
				for id, expire := range expireTracker {
					if expire.Before(time.Now()) {
						delete(pastes, id)
						delete(expireTracker, id)
						log.Printf("id: %s expired\n", id)
					}
				}
			})
			time.Sleep(time.Second)
		}
	}()

	http.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		id := babble()
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		lock.Within(func() {
			pastes[id] = string(body)
			expireTracker[id] = time.Now().Add(expireTime)
		})
		w.Write([]byte(id))
		log.Printf("new paste id: %s content: %s", id, body)
	})
	http.HandleFunc("GET /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("no id supplied"))
			log.Println("id not supplied")
			return
		}

		if paste, ok := pastes[id]; ok {
			w.Write([]byte(paste))
			log.Printf("id: %s found and delivered\n", id)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("paste not found"))
		log.Printf("id: %s not found\n", id)
	})
	http.HandleFunc("DELETE /{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("no id supplied"))
			log.Println("id not supplied")
			return
		}
		delete(pastes, id)
		log.Printf("id: %s deleted\n", id)
	})
	log.Printf("simple paste bin started on port %s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}

var words = loadWords()

func babble() string {
	pieces := []string{}
	for i := 0; i < 2; i++ {
		pieces = append(pieces, words[rand.Int()%len(words)])
	}

	return strings.Join(pieces, "-")
}

func loadWords() (words []string) {
	file, err := os.Open("/usr/share/dict/words")
	if err != nil {
		panic(err)
	}

	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	for _, word := range strings.Split(string(bytes), "\n") {
		if len(word) < 5 && !hasSymbol(word) {
			words = append(words, strings.ToLower(word))
		}
	}
	return words
}

func hasSymbol(s string) bool {
	for _, c := range s {
		if unicode.IsSymbol(c) {
			return true
		}
	}
	return strings.ContainsAny(s, "-'\"")
}
