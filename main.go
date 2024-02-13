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

	"github.com/gorilla/mux"
)

// TODO: refactor to using Go's route parameter stuff. not released yet
// TODO: custom expire times, maybe as a flag

var port = flag.String("port", "4242", "listen port")

var pastes = map[string]string{}

var expireTime = 2 * time.Minute

var expireTracker = map[string]time.Time{}

var lock sync.Mutex

func main() {

	go func() {
		for {
			lock.Lock()
			for id, expire := range expireTracker {
				if expire.Before(time.Now()) {
					delete(pastes, id)
					delete(expireTracker, id)
					log.Printf("id: %s expired\n", id)
				}
			}
			lock.Unlock()
			time.Sleep(time.Second)
		}
	}()

	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("post only"))
			return
		}
		id := babble()
		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		lock.Lock()
		pastes[id] = string(body)
		expireTracker[id] = time.Now().Add(expireTime)
		lock.Unlock()
		w.Write([]byte(id))
		log.Printf("new paste id: %s content: %s", id, body)
	})
	r.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, ok := vars["id"]
		if !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("no id supplied"))
			log.Println("id not supplied")
			return
		}

		if r.Method == http.MethodGet {
			if paste, ok := pastes[id]; ok {
				w.Write([]byte(paste))
				log.Printf("id: %s found and delivered\n", id)
				return
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("paste not found"))
				log.Printf("id: %s not found\n", id)
				return
			}
		}

		if r.Method == http.MethodDelete {
			delete(pastes, id)
			log.Printf("id: %s deleted\n", id)
			return
		}
	})
	log.Printf("simple paste bin started on port %s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, r))
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
