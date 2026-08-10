package main

import (
	"os"
	"reverseproxy/application"
)

func getInItFlag() bool {
	return os.Getenv("INIT") == "1"
}

func main() {
	isNeedToInitilize := getInItFlag()
	app := application.InitializeApplication(isNeedToInitilize)
	app.StartProxy()
}
