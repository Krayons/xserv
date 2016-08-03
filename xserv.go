package main

import (
    "encoding/json"
    "os"
    "fmt"
    "net/http"
    "gopkg.in/tylerb/graceful.v1"
    "github.com/julienschmidt/httprouter"
    "github.com/codegangsta/negroni"
    "io/ioutil"
    "html/template"
    "strings"
    "encoding/base64"
    "bytes"
    "time"
    "sort"
    "./xfile"
)

type TemplatePage struct {
    GenerationTime  time.Duration
    Files           []xfile.DownloadFile
}

type Configuration struct {
    Download_path   string `json:"download_path"`
    Frontend_path   string `json:"frontend_path"`
    Port            string `json:"port"`
    Username        string `json:"username"`
    Password        string `json:"password"`
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
    n := negroni.New(
        negroni.NewRecovery(),
        negroni.NewLogger(),
        negroni.NewStatic(http.Dir(Config.Frontend_path)),
    )
    n.UseHandler(router)
    graceful.Run(":" + Config.Port, 0, n)
}

func RedirectHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    http.Redirect(rw, r, "/browse/", 301)
}

func FileHandler(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
    current_path := Config.Download_path + p[0].Value[1:]
    http.ServeFile(rw, r, current_path)
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
    if (ok) {
        if (val[0] == "date"){
            sort.Stable(xfile.DscDate(download_files))
        }
        if (val[0] == "size"){
            sort.Stable(xfile.AscSize(download_files))
        }
    } else{
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