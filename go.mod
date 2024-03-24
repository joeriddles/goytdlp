module goytdlp

go 1.22.1

// See https://stackoverflow.com/a/72312461 for replacing with a fork
// `go get -d -u github.com/joeriddles/goutubedl@<commit_id>`
// replace github.com/wader/goutubedl => github.com/joeriddles/goutubedl v0.0.0-20240324045750-c44acac69c25

require (
	github.com/u2takey/ffmpeg-go v0.5.0
	github.com/wader/goutubedl v0.0.0-20240314083254-04eb47cbe876
)

require (
	github.com/aws/aws-sdk-go v1.38.20 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/u2takey/go-utils v0.3.1 // indirect
)
