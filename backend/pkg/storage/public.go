package storage

import (
	"bytes"
	"encoding/json"
	"log"
	"net/url"
	"path"
	"rukenshia/frenchwhaling/pkg/wows"
	"rukenshia/frenchwhaling/pkg/wows/api"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type EarnableResource struct {
	Type   wows.Resource
	Amount uint
	Earned uint
}

type StoredShip struct {
	api.ShipStatistics
	Resource EarnableResource
}

type SubscriberPublicData struct {
	AccountID   string
	LastUpdated int64

	Resources []EarnableResource

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

	log.Printf("LoadPublicSubscriberData: key=%s", path.Join("public", parsedURL.Path))

	n, err := svc.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String("frenchwhaling-subscribers"),
		Key:    aws.String(path.Join("public", parsedURL.Path)),
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

func (s *SubscriberPublicData) Save(dataURL string, isNew bool) error {
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

	log.Printf("SubscriberPublicData.Save: key=%s", path.Join("public", parsedURL.Path))
	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String("frenchwhaling-subscribers"),
		Key:    aws.String(path.Join("public", parsedURL.Path)),
		Body:   bytes.NewBuffer(data),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return err
	}

	if isNew {
		log.Printf("SubscriberPublicData.Save: is new subscriber, saving first snapshot")

		_, err = svc.Upload(&s3manager.UploadInput{
			Bucket: aws.String("frenchwhaling-subscribers"),
			Key:    aws.String(path.Join("private", parsedURL.Path)),
			Body:   bytes.NewBuffer(data),
		})
	}

	return nil
}
