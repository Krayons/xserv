package main

import (
    "encoding/json"
    "os"
    "fmt"
    "net/http"
    "github.com/julienschmidt/httprouter"
    "github.com/codegangsta/negroni"
    "io/ioutil"
    "html/template"
    "strings"
    "encoding/base64"
    "bytes"
)

type Configuration struct {
    Download_path   string `json:"download_path"`
    Frontend_path   string `json:"frontend_path"`
    Port            string `json:"port"`
    Username        string `json:"username"`
    Password        string `json:"password"`
}

type DownloadFile struct {
    Name            string  `json:"name"`
    Size            int64   `json:"size"`
    ModTime         int64   `json:"time"`  //Unix time
    IsDir           bool    `json:"isdir"`
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
    router.GET("/", HomeHandler)
    router.GET("/browse/*filepath", BasicAuth(DownloadHandler, user, pass))
    n := negroni.New(
        negroni.NewRecovery(),
        negroni.NewLogger(),
        negroni.NewStatic(http.Dir(Config.Frontend_path)),
    )
    n.UseHandler(router)
    n.Run(":" + Config.Port)
}

func BasicAuth(h httprouter.Handle, user, pass []byte) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
        const basicAuthPrefix string = "Basic "

        // Get the Basic Authentication credentials
        auth := r.Header.Get("Authorization")
        if strings.HasPrefix(auth, basicAuthPrefix) {
            // Check credentials
            payload, err := base64.StdEncoding.DecodeString(auth[len(basicAuthPrefix):])
            if err == nil {
                pair := bytes.SplitN(payload, []byte(":"), 2)
                if len(pair) == 2 &&
                    bytes.Equal(pair[0], user) &&
                    bytes.Equal(pair[1], pass) {

                    // Delegate request to the given handle
                    h(w, r, ps)
                    return
                }
            }
        }

        // Request Basic Authentication otherwise
        w.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
        http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
    }
}

func HomeHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    tmpl, err := template.ParseFiles("./index_template.html")
    if err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := tmpl.Execute(rw, DownloadFile{}); err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
    }
}

func DownloadHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    current_path := Config.Download_path + p[0].Value[1:]    
    files, _ := ioutil.ReadDir(current_path)
    download_files := make([]DownloadFile, len(files))
    var count int = 0
    for _, f := range files {
        var dl = DownloadFile{f.Name(), f.Size(), f.ModTime().Unix(), f.IsDir()}
        download_files[count] = dl
        count++
    }

    //The sorting hat
    query := r.URL.Query()
    if val, ok := query["sort"]; ok{
        fmt.Println(val[0]);
    }

    tmpl, err := template.ParseFiles("./downloads_template.html")
    if err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := tmpl.Execute(rw, download_files); err != nil {
        http.Error(rw, err.Error(), http.StatusInternalServerError)
    }
}