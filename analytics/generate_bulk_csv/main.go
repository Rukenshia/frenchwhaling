package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/charlievieth/fs"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	inFile := flag.String("in", "objects.json", "JSON file of all files that exist on the button (array of Key)")
	outFile := flag.String("out", "last_versions.csv", "output file")
	bucket := flag.String("bucket", "frenchwhaling-subscribers", "s3 bucket name")
	mode := flag.String("mode", "glacier", "mode. either glacier or all to output either all oldest versions or just versions stored in glacier")

	flag.Parse()

	if inFile == nil || outFile == nil {
		log.Printf("in and out are mandatory flags.")
	}

	data, err := ioutil.ReadFile(*inFile)
	if err != nil {
		log.Fatalf("Could not read input: %v", err)
	}

	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		log.Fatalf("Could not parse input: %v", err)
	}

	s3Client := getS3Client()

	output, err := fs.Create(*outFile)
	if err != nil {
		log.Fatalf("Could not open output file: %v", err)
	}
	defer output.Close()

	for _, key := range keys {
		v, err := getOldestVersion(s3Client, *bucket, key)
		if err != nil {
			log.Fatalf("Could not get latest version: %v", err)
		}

		if *mode == "glacier" && v.StorageType != "GLACIER" {
			log.Printf("%s: skipped, glacier mode with non glacier version", key)
			continue
		}

		log.Printf("%s: %s (%v)", key, v.ID, v.Date)
		if _, err := output.WriteString(fmt.Sprintf("%s,%s,%s\n", *bucket, key, v.ID)); err != nil {
			log.Fatalf("Could not write to output file: %v", err)
		}
	}
}

type Version struct {
	ID          string
	Date        *time.Time
	StorageType string
}

func getOldestVersion(client *s3.S3, bucket, key string) (*Version, error) {
	out, err := client.ListObjectVersions(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	v := out.Versions[len(out.Versions)-1]

	return &Version{
		ID:          *v.VersionId,
		StorageType: *v.StorageClass,
		Date:        v.LastModified,
	}, nil
}

func getS3Client() *s3.S3 {
	s, _ := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-central-1"),
		},
	})
	return s3.New(s)
}
