package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func handler(w http.ResponseWriter, r *http.Request, svc *s3.S3, bucket string) {
	key := r.URL.Path
	if key == "/" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	obj, err := svc.GetObject(input)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}
	w.Header().Set("Content-Type", *obj.ContentType)
	buf := new(bytes.Buffer)
	buf.ReadFrom(obj.Body)
	w.Write(buf.Bytes())

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
	bucket := envOrDefault("S3PROXY_BUCKET", "")

	if bucket == "" {
		log.Fatal("You need to provide S3PROXY_BUCKET")
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := s3.New(sess)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, svc, bucket)
	})

	fmt.Printf("Listening on :%s \n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
