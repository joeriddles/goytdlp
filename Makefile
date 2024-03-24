.PHONY: run
run:
	go run main.go

.PHONY: test
test:
	curl -v -X POST -d "url=https://www.youtube.com/watch?v=rwBrd5XZPAc" http://localhost:8080/download --output song.mp3
