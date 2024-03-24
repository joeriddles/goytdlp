# goytdlp

A basic Go web app to convert YouTube videos into MP3 files.

To run:
```shell
go run main.go
```

To run the Dockerfile:
```shell
docker build -t goytdlp:latest .
docker run --rm -p '8080:8080' goytdlp:latest
```
