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

	"github.com/GoIncremental/negroni-sessions/redisstore"
	sessions "github.com/goincremental/negroni-sessions"
	"github.com/julienschmidt/httprouter"
	"github.com/krayons/xserv/xfile"
	"github.com/urfave/negroni"
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

func main() {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config = configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("error:", err)
	}
	router := httprouter.New()
	router.GET("/", redirectHandler)
	router.GET("/browse/*filepath", loginOnly(downloadHandler))
	router.GET("/downloads/*filepath", fileHandler)
	router.GET("/login/", loginHandler)
	router.GET("/logout/", logoutHandler)
	router.POST("/login/", loginHandler)
	if err != nil {
		fmt.Println("error:", err)
	}
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
		negroni.NewStatic(http.Dir(config.FrontendPath)),
	)
	rStore, err := redisstore.New(10, "tcp", "localhost:6379", "", []byte("supersecret"))
	n.Use(sessions.Sessions("my_session", rStore))
	n.UseHandler(router)
	graceful.Run(":"+config.Port, 0, n)
}

func logoutHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	session := sessions.GetSession(r)
	session.Clear()
	http.Redirect(rw, r, "/login/", 302)
}
func redirectHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	http.Redirect(rw, r, "/browse/", 301)
}

func loginHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if r.Method == "POST" {
		r.ParseForm()
		if (r.Form["username"][0] == config.Username) && (r.Form["password"][0] == config.Password) {
			session := sessions.GetSession(r)
			session.Set("loggedin", "yes")
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

func fileHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	currentPath := config.DownloadPath + p[0].Value[1:]
	http.ServeFile(rw, r, currentPath)
}

func loginOnly(h httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		session := sessions.GetSession(r)
		if session.Get("loggedin") != nil {
			h(w, r, ps)
			return
		}
		http.Redirect(w, r, "/login/?redirect="+r.RequestURI, 302)
		return
	}
}

func downloadHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	fmt.Println("hi")
	start := time.Now()
	currentPath := config.DownloadPath + p[0].Value[1:]
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
