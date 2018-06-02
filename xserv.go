package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/krayons/xserv/xfile"
)

type templatePage struct {
	GenerationTime time.Duration
	Files          []xfile.DownloadFile
}

type configuration struct {
	DownloadPath string `json:"downloadPath"`
	FrontendPath string `json:"frontendPath"`
	Port         string `json:"port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var config configuration

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config = configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	router.HandleFunc("/", redirectHandler)
	router.PathPrefix("/browse").HandlerFunc(loggedinOnly(downloadHandler))
	router.PathPrefix("/downloads").HandlerFunc(fileHandler)
	router.HandleFunc("/login/", loginHandler)
	router.HandleFunc("/logout/", logoutHandler)
	err = http.ListenAndServe(":"+config.Port, router)
	fmt.Println(err)
}

func logoutHandler(rw http.ResponseWriter, r *http.Request) {
	store.MaxAge(-1)
	http.Redirect(rw, r, "/login/", 302)
}
func redirectHandler(rw http.ResponseWriter, r *http.Request) {
	http.Redirect(rw, r, "/browse/", http.StatusPermanentRedirect)
}

func loginHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		if (r.Form["username"][0] == config.Username) && (r.Form["password"][0] == config.Password) {
			session, _ := store.Get(r, "session")
			session.Values["logged_in"] = true
			session.Save(r, rw)
			http.Redirect(rw, r, "/browse/", 302)
		}

	}
	tmpl, err := template.ParseFiles("./templates/login_template.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	templatePage := templatePage{}
	if err := tmpl.Execute(rw, templatePage); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func fileHandler(rw http.ResponseWriter, r *http.Request) {
	re := regexp.MustCompile(`downloads\/(.*)`)
	p := re.FindAllStringSubmatch(r.URL.Path, 1)
	var currentPath string
	if p != nil {
		currentPath = config.DownloadPath + p[0][1]
	} else {
		currentPath = config.DownloadPath
	}
	http.ServeFile(rw, r, currentPath)
}

func loggedinOnly(f func(w http.ResponseWriter, r *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		isLoggedIn := session.Values["logged_in"]
		fmt.Println(isLoggedIn)
		if isLoggedIn != nil {
			f(w, r)
		} else {
			http.Redirect(w, r, "/login/?redirect="+r.RequestURI, http.StatusTemporaryRedirect)
		}
	}
}

func downloadHandler(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()
	re := regexp.MustCompile(`browse\/(.*)`)
	p := re.FindAllStringSubmatch(r.URL.Path, 1)
	var currentPath string
	if p != nil {
		currentPath = config.DownloadPath + p[0][1]
	} else {
		currentPath = config.DownloadPath
	}
	files, _ := ioutil.ReadDir(currentPath)
	downloadFiles := make([]xfile.DownloadFile, len(files))
	var count int
	for _, f := range files {
		var dl = xfile.DownloadFile{f.Name(), f.Size(), f.ModTime().Unix(), f.IsDir()}
		downloadFiles[count] = dl
		count++
	}

	//The sorting hat
	query := r.URL.Query()
	val, ok := query["sort"]
	if ok {
		if val[0] == "date" {
			sort.Stable(xfile.DscDate(downloadFiles))
		}
		if val[0] == "size" {
			sort.Stable(xfile.AscSize(downloadFiles))
		}
	} else {
		sort.Stable(xfile.AcsName(downloadFiles))
	}

	tmpl, err := template.ParseFiles("./templates/downloads_template.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	templatePage := templatePage{time.Since(start), downloadFiles}
	if err := tmpl.Execute(rw, templatePage); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
