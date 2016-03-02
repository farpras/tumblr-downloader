package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb"
)

var (
	gfyRequest = "https://gfycat.com/cajax/get/%s"
)

// A File contains information on a particular tumblr URL, as well as the user where the URL was found.
type File struct {
	User          string
	URL           string
	UnixTimestamp int64
	ProgressBar   *pb.ProgressBar
}

// Download downloads a file specified in the file's URL.
func (f File) Download() {
	var resp *http.Response
	for {
		resp2, err := http.Get(f.URL)
		if err != nil {
			log.Println(err)
		} else {
			resp = resp2
			break
		}
	}
	defer resp.Body.Close()

	pic, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ReadAll:", err)
	}
	filename := path.Join(downloadDirectory, f.User, path.Base(f.URL))

	err = ioutil.WriteFile(filename, pic, 0644)
	if err != nil {
		log.Fatal("WriteFile:", err)
	}

	err = os.Chtimes(filename, time.Now(), time.Unix(f.UnixTimestamp, 0))
	if err != nil {
		log.Println(err)
	}

	f.ProgressBar.Increment()
	atomic.AddUint64(&totalDownloaded, 1)
	atomic.AddUint64(&totalSizeDownloaded, uint64(len(pic)))

}

// String is the standard method for the Stringer interface.
func (f File) String() string {
	date := time.Unix(f.UnixTimestamp, 0)
	return f.User + " - " + date.Format("2006-01-02 15:04:05") + " - " + path.Base(f.URL)
}

// Gfy houses the Gfycat response.
type Gfy struct {
	GfyItem struct {
		Mp4Url  string `json:"mp4Url"`
		WebmURL string `json:"webmUrl"`
	} `json:"gfyItem"`
}

// GetGfycatURL gets the appropriate Gfycat URL for download, from a "normal" link.
func GetGfycatURL(slug string) string {
	gfyURL := fmt.Sprintf(gfyRequest, slug)

	var resp *http.Response
	for {
		resp2, err := http.Get(gfyURL)
		if err != nil {
			log.Println("GetGfycatURL: ", err)
		} else {
			resp = resp2
			break
		}
	}
	defer resp.Body.Close()

	gfyData, _ := ioutil.ReadAll(resp.Body)

	var gfy Gfy

	err := json.Unmarshal(gfyData, &gfy)
	if err != nil {
		log.Println("GfycatUnmarshal: ", err)
	}

	return gfy.GfyItem.Mp4Url
}