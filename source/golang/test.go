package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

type Page struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Url   string `json:"url"`
}

func (p Page) toString() string {
    return toJson(p)
}

func toJson(p interface{}) string {
    bytes, err := json.Marshal(p)
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    return string(bytes)
}

func main() {

    pages := getPages()
    for _, p := range pages {
        fmt.Println(string(p.Title))
    }

    fmt.Println(toJson(pages))
}

func getPages() []Page {
    raw, err := ioutil.ReadFile("./public/json/examplepages.json")
    if err != nil {
        fmt.Println(err.Error())
        os.Exit(1)
    }

    var c []Page
    json.Unmarshal(raw, &c)
    return c
}
