package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Service  *s3.S3
	bucketName string
)

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	key := r.URL.Path
	if key == "/" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}
	obj, err := s3Service.GetObject(input)
	if err != nil {
		log.Printf("Error while getting %q: %s\n", key, err.Error())
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	defer obj.Body.Close()

	w.Header().Set("Content-Type", *obj.ContentType)

	// Directly copy all bytes from the S3 object into the HTTP reponse
	io.Copy(w, obj.Body)
}

func envOrDefault(name string, defaultValue string) string {
	if os.Getenv(name) != "" {
		return os.Getenv(name)
	} else {
		return defaultValue
	}
}

func main() {
	region := envOrDefault("S3PROXY_REGION", "eu-central-1")
	port := envOrDefault("S3PROXY_PORT", "3000")
	bucketName = envOrDefault("S3PROXY_BUCKET", "")

	if bucketName == "" {
		log.Fatal("You need to provide S3PROXY_BUCKET")
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	s3Service = s3.New(sess)

	http.HandleFunc("/", handler)

	log.Printf("Listening on :%s \n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
