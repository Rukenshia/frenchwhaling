package storage

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"rukenshia/frenchwhaling/pkg/wows"
	"rukenshia/frenchwhaling/pkg/wows/api"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type EarnabledResource struct {
	Resource wows.Resource
	Amount   uint
	Earned   bool
}

type StoredShip struct {
	api.ShipStatistics
	Earnable EarnabledResource
}

type SubscriberPublicData struct {
	AccountID   string
	LastUpdated *time.Time

	Earnable []EarnabledResource

	Ships map[int64]*StoredShip
}

func LoadPublicSubscriberData(dataURL string) (*SubscriberPublicData, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-central-1"),
		},
	})
	if err != nil {
		return nil, err
	}
	svc := s3manager.NewDownloader(sess)

	parsedURL, err := url.Parse(dataURL)
	if err != nil {
		return nil, err
	}

	buf := &aws.WriteAtBuffer{}

	log.Printf("LoadPublicSubscriberData: key=%s", strings.Replace(parsedURL.Path, "/data", "/public", 1))

	n, err := svc.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String("frenchwhaling-subscribers"),
		Key:    aws.String(strings.Replace(parsedURL.Path, "/data", "/public", 1)),
	})
	if err != nil {
		return nil, err
	}
	log.Printf("LoadPublicSubscriberData: downloaded %d bytes", n)

	var data SubscriberPublicData
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (s *SubscriberPublicData) Save(dataURL string) error {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("eu-central-1"),
		},
	})
	if err != nil {
		return err
	}
	svc := s3manager.NewUploader(sess)

	parsedURL, err := url.Parse(dataURL)
	if err != nil {
		return err
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	log.Printf("SubscriberPublicData.Save: key=%s", strings.Replace(parsedURL.Path, "/data", "/public", 1))

	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String("frenchwhaling-subscribers"),
		Key:    aws.String(strings.Replace(parsedURL.Path, "/data", "/public", 1)),
		Body:   bytes.NewBuffer(data),
	})

	return err
}
