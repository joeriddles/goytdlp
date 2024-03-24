package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/wader/goutubedl"
)

var indexTemplate *template.Template

func main() {
	http.HandleFunc("GET /{$}", index)
	http.HandleFunc("POST /download", downloadVideo)

	fp := path.Join("templates", "index.html")
	var err error
	indexTemplate, err = template.ParseFiles(fp)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":8080"
	fmt.Printf("server starting on %v\n", addr)
	err = http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	indexTemplate.Execute(w, nil)
}

func downloadVideo(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	fmt.Printf("Downloading %v\n", url)

	metadata, err := goutubedl.New(context.Background(), url, goutubedl.Options{
		DebugLog: &printer{},
		// ExtractAudio: true,
		// AudioFormat:  "mp3",
	})
	if err != nil {
		writeError(err, w)
		return
	}

	result, err := metadata.Download(context.Background(), "ba") // best audio
	if err != nil {
		writeError(err, w)
		return
	}
	defer result.Close()

	tempPath, err := os.MkdirTemp("", "main")
	defer os.RemoveAll(tempPath)
	if err != nil {
		writeError(err, w)
		return
	}

	ytdlpFilePath := path.Join(tempPath, "download.webm")
	ytdlpFile, err := os.Create(ytdlpFilePath)
	if err != nil {
		writeError(err, w)
		return
	}

	resultBytes, err := io.ReadAll(result)
	if err != nil {
		writeError(err, w)
		return
	}

	_, err = ytdlpFile.Write(resultBytes)
	if err != nil {
		writeError(err, w)
		return
	}

	err = ytdlpFile.Close()
	if err != nil {
		writeError(err, w)
		return
	}

	mp3FilePath := ytdlpFilePath + ".mp3"
	err = ffmpeg.Input(ytdlpFilePath).
		Audio().
		Output(mp3FilePath, ffmpeg.KwArgs{"b:a": "192k"}).
		OverWriteOutput().
		ErrorToStdOut().
		Run()
	if err != nil {
		writeError(err, w)
		return
	}

	mp3File, err := os.Open(mp3FilePath)
	if err != nil {
		writeError(err, w)
		return
	}

	w.Header().Set("Content-Type", "audio/mpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("filename=\"%v.mp3\"", metadata.Info.Title))

	io.Copy(w, mp3File)
}

func writeError(err error, w http.ResponseWriter) {
	log.Print(err.Error())
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(400)
	io.WriteString(w, err.Error())
}

var _ goutubedl.Printer = &printer{}

type printer struct{}

func (printer) Print(v ...interface{}) {
	fmt.Print(v...)
	fmt.Println()
}
