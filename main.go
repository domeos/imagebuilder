package main

import (
    "imagebuilder/build"
    "os"
)

func main() {
    codeType := os.Getenv("TYPE")
    build.RunOnType(codeType)
}
