package main

import "github.com/dathan/go-upload-bridge-pinata/pkg/upload"

func main() {
	upload.SetUpServer(":8080")
}
