package main

import (
	"fmt"
	"path/filepath"
	"os"
	"log"
	"sync"
	"time"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const DATE_PATTERN = "20060102"
const COMMAND = "backup-to-s3"

var (
	wg sync.WaitGroup
)

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s path bucket host\n", COMMAND)
		return
	}

	path := os.Args[1]
	bucket := os.Args[2]
	host := os.Args[3]

	now := time.Now()

	files, _ := filepath.Glob(filepath.Join(path, "*.gz"))
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			upload(bucket, host, now, file)
			wg.Done()
		}(file)
	}

	wg.Wait()

	remove(bucket, host, now)
}

func upload(bucket string, host string, now time.Time, path string) {
	filename := filepath.Base(path)

	fp, err := os.Open(path)
	if err != nil {
		log.Fatal("failed to open", path)
	}
	defer fp.Close()

	uploadPath := fmt.Sprintf("%s/%s/%s", host, now.Format(DATE_PATTERN), filename)
	log.Printf("UPLOAD FILE: %s ===> %s\n", path, uploadPath)

	uploader := s3manager.NewUploader(s3Session())
	_, err = uploader.Upload(&s3manager.UploadInput{
		Body: fp,
		Bucket: aws.String(bucket),
		Key: aws.String(uploadPath),
	})
	if err != nil {
		log.Fatal("failed to upload", err)
	}
}

func remove(bucket string, host string, now time.Time) {
	deleteTime := now.Add(time.Duration(-30 * 24) * time.Hour)
	deletePath := fmt.Sprintf("%s/%s/", host, deleteTime.Format(DATE_PATTERN))

	svc := s3.New(s3Session())
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(deletePath),
	})
	if err != nil {
		log.Fatal("failed to delete", err)
	}

	log.Println("REMOVE FILE:", deletePath)
}

func s3Session() *session.Session {
	sess, err := session.NewSession()
	if err != nil {
		log.Fatal("Error creating session ", err)
	}

	return sess
}
