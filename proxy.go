package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

var (
	s3Service  *s3.S3
	bucketName string
	cachePath  string
)

// getFile checks if we have a local copy otherwise downloads from S3
func getFile(key string) (FileWrapper, error) {
	if cachePath != "" {
		log.Debug("Trying to get file from cache")
		obj, err := getFileFromCache(key)

		// Directly return file from Cache if we didn't got an error
		if err == nil {
			log.Info("Returning cached file")
			return obj, nil
		} else {
			log.Debug(err)
		}
	}

	obj, err := getFileFromBucket(key)
	if err != nil {
		return FileWrapper{}, err
	}

	log.Debug("Returning file from Bucket")
	return obj, nil

}

func getFileFromCache(key string) (FileWrapper, error) {
	filePath := filepath.Join(cachePath, key)

	if fileStat, err := os.Stat(filePath); err == nil {
		// file in cache. check expire
		headRequest, err := s3Service.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(key),
		})

		if err != nil {
			// We have a local file, but HeadObject returned an error, so we can
			// assume that the file no longer exists in the bucket
			os.Remove(filePath)
			log.Debug("Deleting local file")
			return FileWrapper{}, err
		}

		if fileStat.ModTime().Before(*headRequest.LastModified) {
			// Our file is older than the one in the bucket
			os.Remove(filePath)
			return FileWrapper{}, errors.New("file not up to date")
		}

		fh, err := os.Open(filePath)
		if err != nil {
			// Couldn't open cached file
			return FileWrapper{}, err
		}

		return FileWrapper{
			File:       fh,
			HeadOutput: headRequest,
			GetOutput:  nil,
		}, nil

	} else {
		// File not in cache or otherwise not accessible
		return FileWrapper{}, err
	}
}

func getFileFromBucket(key string) (FileWrapper, error) {
	log.Info("Getting file from Bucket")

	obj, err := s3Service.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Errorf("Error while getting %q from S3: %s\n", key, err.Error())
		return FileWrapper{}, err
	}

	s3File := FileWrapper{
		File:       nil,
		HeadOutput: nil,
		GetOutput:  obj,
	}

	if cachePath != "" {
		path, err := saveFileToCache(key, obj)
		if err != nil {
			// We couldn't save the file to the cache but still return the Get response from S3
			log.Error(err)
			return s3File, nil
		}

		fh, _ := os.Open(path)
		return FileWrapper{
			File:       fh,
			HeadOutput: nil,
			GetOutput:  obj,
		}, nil

	}

	return s3File, nil
}

// createWithFolders creates the full nested directory structure and then creates the requested file
func createWithFolders(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func saveFileToCache(key string, obj *s3.GetObjectOutput) (string, error) {
	log.Debug("Saving file to cache")
	filePath := filepath.Join(cachePath, key)

	outFile, err := createWithFolders(filePath)
	if err != nil {
		log.Error("Couldn't create cache dir")
		return "", err
	}
	defer outFile.Close()

	io.Copy(outFile, obj.Body)

	return filePath, nil

}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	key := r.URL.Path
	if key == "/" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	log.WithFields(log.Fields{
		"key": key,
	}).Info("Got a request")

	obj, err := getFile(key)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	// Set correct ContentType
	w.Header().Set("Content-Type", obj.GetContentType())

	// Check for additional metadata
	metadata := obj.GetMetadata()
	if len(metadata) > 0 {
		for k, v := range metadata {
			w.Header().Set(k, *v)
		}
	}

	// Directly copy all bytes from the S3 object into the HTTP reponse
	io.Copy(w, obj.GetContent())
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
	cachePath = envOrDefault("S3PROXY_CACHE", "")
	logLevel := envOrDefault("S3PROXY_LOGGING", "WARN")

	l, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Error("Unknown loglevel provided. Defaulting to WARN")
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(l)
	}

	if bucketName == "" {
		log.Fatal("You need to provide S3PROXY_BUCKET")
	}

	if cachePath != "" {
		// Check if we have write access to the cache directory
		testPath := filepath.Join(cachePath, ".testfile")
		file, err := createWithFolders(testPath)
		if err != nil {
			log.Fatal("No write access to the cache dir")
		}
		defer file.Close()

	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	s3Service = s3.New(sess)

	http.HandleFunc("/", handler)

	log.Info("Listening on :%s \n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
