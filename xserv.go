package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	graceful "gopkg.in/tylerb/graceful.v1"

	"./xfile"
	"github.com/GoIncremental/negroni-sessions/redisstore"
	sessions "github.com/goincremental/negroni-sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

type TemplatePage struct {
	GenerationTime time.Duration
	Files          []xfile.DownloadFile
}

type Configuration struct {
	Download_path string `json:"download_path"`
	Frontend_path string `json:"frontend_path"`
	Port          string `json:"port"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

var Config Configuration

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	Config = Configuration{}
	err := decoder.Decode(&Config)
	if err != nil {
		fmt.Println("error:", err)
	}
	user := []byte(Config.Username)
	pass := []byte(Config.Password)
	router := httprouter.New()
	router.GET("/", RedirectHandler)
	router.GET("/browse/*filepath", BasicAuth(DownloadHandler, user, pass))
	router.GET("/downloads/*filepath", FileHandler)
	router.GET("/login/", loginHandler)
	router.POST("/login/", loginHandler)
	if err != nil {
		fmt.Println("error:", err)
	}
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(Config.Frontend_path)),
	)
	rStore, err := redisstore.New(10, "tcp", "localhost:6379", "", []byte("secret123"))
	n.Use(sessions.Sessions("my_session", rStore))
	n.UseHandler(router)
	graceful.Run(":"+Config.Port, 0, n)
}

func RedirectHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.Redirect(rw, r, "/browse/", 301)
}

func loginHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tmpl, err := template.ParseFiles("./login_template.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	templatePage := TemplatePage{}
	if err := tmpl.Execute(rw, templatePage); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func FileHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	current_path := Config.Download_path + p[0].Value[1:]
	http.ServeFile(rw, r, current_path)
}

func BasicAuth(h httprouter.Handle, user, pass []byte) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		session := sessions.GetSession(r)
		if session.Get("loggedin") != nil {
			h(w, r, ps)
		}
		h(w, r, ps)
		return
	}
}

func DownloadHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	start := time.Now()
	current_path := Config.Download_path + p[0].Value[1:]
	files, _ := ioutil.ReadDir(current_path)
	download_files := make([]xfile.DownloadFile, len(files))
	var count int = 0
	for _, f := range files {
		var dl = xfile.DownloadFile{f.Name(), f.Size(), f.ModTime().Unix(), f.IsDir()}
		download_files[count] = dl
		count++
	}

	//The sorting hat
	query := r.URL.Query()
	val, ok := query["sort"]
	if ok {
		if val[0] == "date" {
			sort.Stable(xfile.DscDate(download_files))
		}
		if val[0] == "size" {
			sort.Stable(xfile.AscSize(download_files))
		}
	} else {
		sort.Stable(xfile.AcsName(download_files))
	}

	tmpl, err := template.ParseFiles("./downloads_template.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	templatePage := TemplatePage{time.Since(start), download_files}
	if err := tmpl.Execute(rw, templatePage); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
