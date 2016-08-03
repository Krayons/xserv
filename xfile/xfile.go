package xfile

type DownloadFile struct {
    Name            string  `json:"name"`
    Size            int64   `json:"size"`
    ModTime         int64   `json:"time"`  //Unix time
    IsDir           bool    `json:"isdir"`
}


func (slice AscDate) BothDir(i, j int) bool {
    if (slice[i].IsDir == slice[j].IsDir){
        return true
    }
    return false
}

// Sort by AscDate
type AscDate []DownloadFile

func (slice AscDate) Len() int {
    return len(slice)
}

func (slice AscDate) Less(i, j int) bool {
    if (slice.BothDir(i, j)){
        return slice[i].ModTime < slice[j].ModTime
    }
    if (slice[i].IsDir){
        return true
    }
    return false
}

func (slice AscDate) Swap(i, j int){
    slice[i], slice[j] = slice[j], slice[i]
}

// Sort by Date decending
type DscDate []DownloadFile

func (slice DscDate) BothDir(i, j int) bool {
    if (slice[i].IsDir == slice[j].IsDir){
        return true
    }
    return false
}

func (slice DscDate) Len() int {
    return len(slice)
}

func (slice DscDate) Less(i, j int) bool {
    if (slice.BothDir(i, j)){
        return slice[i].ModTime > slice[j].ModTime
    }
    if (slice[i].IsDir){
        return true
    }
    return false
}

func (slice DscDate) Swap(i, j int){
    slice[i], slice[j] = slice[j], slice[i]
}


// Sort Acessending Size
type AscSize []DownloadFile

func (slice AscSize) Len() int {
    return len(slice)
}

func (slice AscSize) Less(i, j int) bool {
    return slice[i].Size < slice[j].Size
}

func (slice AscSize) Swap(i, j int){
    slice[i], slice[j] = slice[j], slice[i]
}

// Sort Ascendingly by name
type AcsName []DownloadFile

func (slice AcsName) BothDir(i, j int) bool {
    if (slice[i].IsDir == slice[j].IsDir){
        return true
    }
    return false
}

func (slice AcsName) Len() int {
    return len(slice)
}

func (slice AcsName) Less(i, j int) bool {
    if (slice.BothDir(i, j)){
        return slice[i].Name < slice[j].Name
    }
    return slice[i].IsDir
}

func (slice AcsName) Swap(i, j int){
    slice[i], slice[j] = slice[j], slice[i]
}