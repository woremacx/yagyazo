package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var portNumber = flag.String("port", "8080", "port number.")

var dir string

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello "+r.URL.Path)
}

func makeBasename() string {
	basename := strconv.FormatInt(time.Now().UnixNano(), 10) + os.Getenv("SECRET")

	h := sha1.New()
	h.Write([]byte(basename))
	bs := h.Sum(nil)

	dst := make([]byte, hex.EncodedLen(len(bs)))
	hex.Encode(dst, bs)

	return string(dst)
}

func genDirname(s string) string {
	dst := path.Join("images", string(s[0]), string(s[0:2]))
	return dst
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	basepart := makeBasename()
	imagedir := genDirname(basepart)

	imagedirpath := path.Join(dir, imagedir)
	if err := os.MkdirAll(imagedirpath, 0755); err != nil && !os.IsExist(err) {
		fmt.Fprintln(w, err)
		return
	}

	file, _, err := r.FormFile("imagedata")
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer file.Close()

	basename := basepart + ".png"
	imagefile := path.Join(imagedirpath, basename)
	out, err := os.Create(imagefile)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	url := fmt.Sprintf("%s/images/%s", os.Getenv("URLBASE"), basename)
	fmt.Println(url)

	fmt.Fprintf(w, url)
}

func imagesHandler(w http.ResponseWriter, r *http.Request) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	imagefile := path.Join(dir, r.URL.Path)

	http.ServeFile(w, r, imagefile)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Error cannot get Getwd")
		return
	}
	dir = pwd

	flag.Parse()
	log.Println("listen:" + *portNumber)

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/images/", imagesHandler)
	http.HandleFunc("/upload.cgi", uploadHandler)
	http.ListenAndServe(":"+*portNumber, nil)
}
