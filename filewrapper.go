package main

import (
	"io"
	"os"

	"github.com/aws/aws-sdk-go/service/s3"
)

// FileWrapper wraps either a local file or an reponse from S3
// It either contains a pointer to a local file and the reponse from a HeadObject request
// or both of these are nil and it only contains an GetObject request
type FileWrapper struct {
	File       *os.File
	GetOutput  *s3.GetObjectOutput
	HeadOutput *s3.HeadObjectOutput
}

func (obj *FileWrapper) GetContent() io.Reader {
	if obj.File != nil {
		return obj.File
	} else {
		return obj.GetOutput.Body
	}
}

func (obj *FileWrapper) GetContentType() string {
	if obj.GetOutput != nil {
		return *obj.GetOutput.ContentType
	} else {
		return *obj.HeadOutput.ContentType
	}
}

func (obj *FileWrapper) GetMetadata() map[string]*string {
	if obj.GetOutput != nil {
		return obj.GetOutput.Metadata
	} else {
		return obj.HeadOutput.Metadata
	}
}
