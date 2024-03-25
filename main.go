package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/wader/goutubedl"
)

//go:embed templates
var templates embed.FS

var indexTemplate *template.Template
var downloadTemplate *template.Template

type DownloadTemplateData struct {
	FileName string
	ImageUrl string
}

func main() {
	http.HandleFunc("GET /{$}", index)
	http.HandleFunc("GET /{filename}/{$}", viewMedia)
	http.HandleFunc("GET /media/{filename}/{$}", getMedia)
	http.HandleFunc("POST /download/{$}", downloadVideo)

	indexTemplate = template.Must(template.
		New("index.html").
		ParseFS(templates, "templates/base.html", "templates/index.html"))
	downloadTemplate = template.Must(template.
		New("download.html").
		ParseFS(templates, "templates/base.html", "templates/download.html"))

	addr := ":8080"
	fmt.Printf("server starting on http://localhost%v\n", addr)
	err := http.ListenAndServe(addr, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	err := indexTemplate.Execute(w, nil)
	if err != nil {
		writeError(err, w)
	}
}

func viewMedia(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	imageFilename := strings.Replace(filename, ".mp3", ".jpg", 1)
	imageUrl := fmt.Sprintf("/media/%v/", imageFilename)
	data := &DownloadTemplateData{FileName: filename, ImageUrl: imageUrl}
	err := downloadTemplate.Execute(w, data)
	if err != nil {
		writeError(err, w)
	}
}

func getMedia(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	filepath := path.Join("media", filename)
	file, err := os.Open(filepath)
	if err != nil {
		log.Print(err.Error())
		w.WriteHeader(404)
		return
	}
	if strings.HasSuffix(filename, ".mp3") {
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", filename))
	}
	io.Copy(w, file)
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

	mp3FileName := metadata.Info.Title + ".mp3"
	mp3FileName = clearString(mp3FileName)
	mp3FilePath := path.Join("media", mp3FileName)
	_, err = os.Stat(mp3FilePath)
	if err != nil {
		_, err = downloadAndConvertVideo(metadata, mp3FileName)
		if err != nil {
			writeError(err, w)
			return
		}
	}

	_, err = downloadThumbnail(metadata, mp3FileName)
	if err != nil {
		writeError(err, w)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/%v/", mp3FileName), http.StatusSeeOther)
}

func downloadThumbnail(metadata goutubedl.Result, mp3Filename string) (string, error) {
	thumbnailUrl := metadata.Info.Thumbnail
	mp3Name := mp3Filename[0:strings.LastIndex(mp3Filename, ".")]
	// fileExtension := mp3Filename[strings.LastIndex(mp3Filename, ".")+1:]
	filename := fmt.Sprintf("%v.%v", mp3Name, "jpg")
	filepath := path.Join("media", filename)
	out, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	resp, err := http.Get(thumbnailUrl)
	if err != nil {
		return "", nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

func downloadAndConvertVideo(metadata goutubedl.Result, filename string) (*os.File, error) {
	result, err := metadata.Download(context.Background(), "ba") // best audio
	if err != nil {
		return nil, err
	}
	defer result.Close()

	tempPath, err := os.MkdirTemp("", "main")
	defer os.RemoveAll(tempPath)
	if err != nil {
		return nil, err
	}

	ytdlpFilePath := path.Join(tempPath, "download.webm")
	ytdlpFile, err := os.Create(ytdlpFilePath)
	if err != nil {
		return nil, err
	}

	resultBytes, err := io.ReadAll(result)
	if err != nil {
		return nil, err
	}

	_, err = ytdlpFile.Write(resultBytes)
	if err != nil {
		return nil, err
	}

	err = ytdlpFile.Close()
	if err != nil {
		return nil, err
	}

	mp3FilePath := path.Join("media", filename)
	err = ffmpeg.Input(ytdlpFilePath).
		Audio().
		Output(mp3FilePath, ffmpeg.KwArgs{"b:a": "192k"}).
		OverWriteOutput().
		ErrorToStdOut().
		Run()
	if err != nil {
		return nil, err
	}

	mp3File, err := os.Open(mp3FilePath)
	if err != nil {
		return nil, err
	}

	return mp3File, nil
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

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-. ]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}
