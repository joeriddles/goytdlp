# goytdlp

A basic Go web app to convert YouTube videos into MP3 files.

To run:
```shell
go run main.go
```

To build locally:
```shell
go build -v -o ./run-app .
```

To build and run the Dockerfile:
```shell
docker build -t goytdlp:latest .
docker run --rm -p '8080:8080' goytdlp:latest
```
