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

	// Upload archive files
	files, _ := filepath.Glob(filepath.Join(path, "*.gz"))
	for _, file := range files {
		wg.Add(1)
		go func(file string) {
			upload(bucket, host, now, file)
			wg.Done()
		}(file)
	}

	wg.Wait()

	// Remove expired archive files
	deleteTime := now.Add(time.Duration(-30 * 24) * time.Hour)
	deletePath := fmt.Sprintf("%s/%s/", host, deleteTime.Format(DATE_PATTERN))
	svc := s3.New(s3Session())

	objects := listObjects(svc, bucket, deletePath)
	if len(objects.Contents) > 0 {
		remove(svc, bucket, objects)
	}

	log.Println("REMOVE FILE:", deletePath)
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

func listObjects(svc *s3.S3, bucket string, deletePath string) *s3.ListObjectsOutput {
	resp, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(deletePath),
	})
	if err != nil {
		log.Fatal("failed to list objects", err)
	}

	return resp
}

func remove(svc *s3.S3, bucket string, objects *s3.ListObjectsOutput) {
	targets := []*s3.ObjectIdentifier{}
	for _, c := range objects.Contents {
		targets = append(targets, &s3.ObjectIdentifier{
			Key: c.Key,
		})
	}

	_, err := svc.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3.Delete{
			Objects: targets,
		},
	})
	if err != nil {
		log.Fatal(err.Error())
		log.Fatal("failed to delete objects")
	}
}

func s3Session() *session.Session {
	// failed to list objectsMissingRegion: could not find region configuration
	//sess, err := session.NewSession()
	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-northeast-1")})
	if err != nil {
		log.Fatal("Error creating session ", err)
	}

	return sess
}
