package main

// This is a dummy main.go file to import dependencies used in the integration tests (./_integration_tests).
import (
	_ "cloud.google.com/go/secretmanager/apiv1"
	_ "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	_ "cloud.google.com/go/storage"
	_ "golang.org/x/oauth2/google"
	_ "golang.org/x/oauth2/jwt"
	_ "google.golang.org/api/option"
)

func main() {

}
