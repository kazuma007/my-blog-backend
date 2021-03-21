package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	S3AK string
	S3SK string
}

type RequestBody struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	File      string `json:"file"`
	Extension string `json:"extension"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestBody := new(RequestBody)
	if err := json.Unmarshal(([]byte)(request.Body), requestBody); err != nil {
		fmt.Println("json unmarshal error", err)
	}

	uuidV4 := uuid.New()
	createdId := uuidV4.String()

	var env Env
	if err := envconfig.Process("", &env); err != nil {
		fmt.Println("envconfig error", err)
		return events.APIGatewayProxyResponse{}, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(env.S3AK, env.S3SK, ""),
		Region:      aws.String("ap-northeast-1"),
	}))

	if err := updateDynamoDB(requestBody, createdId, sess); err != nil {
		fmt.Println("updateDynamoDB error", err)
		return events.APIGatewayProxyResponse{}, nil
	}

	if len(requestBody.File) != 0 {
		if err := updateS3(requestBody, createdId, sess); err != nil {
			fmt.Println("updateS3 error", err)
			return events.APIGatewayProxyResponse{}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
			"Content-Type":                 "application/json",
		},
		Body:       "OK",
		StatusCode: 200,
	}, nil
}

func updateDynamoDB(requestBody *RequestBody, createdId string, sess *session.Session) error {
	svc := dynamodb.New(sess)

	slice := strings.Split(requestBody.Extension, "/")

	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
	time.LoadLocation("Asia/Tokyo")
	t := time.Now()
	layout := "2006/01/02T15:04:05"
	registeredTime := t.Format(layout)

	putParams := &dynamodb.PutItemInput{
		TableName: aws.String("my-blog-t"),
		Item: map[string]*dynamodb.AttributeValue{
			"storage_key": {
				S: aws.String(createdId + "." + slice[1]),
			},
			"filename": {
				S: aws.String(createdId),
			},
			"title": {
				S: aws.String(requestBody.Title),
			},
			"content": {
				S: aws.String(requestBody.Content),
			},
			"registered_time": {
				S: aws.String(registeredTime),
			},
		},
	}

	if _, err := svc.PutItem(putParams); err != nil {
		return err
	}

	return nil
}

func updateS3(requestBody *RequestBody, createdId string, sess *session.Session) error {
	uploader := s3manager.NewUploader(sess)

	data, err := base64.StdEncoding.DecodeString(requestBody.File[strings.IndexByte(requestBody.File, ',')+1:])

	if err != nil {
		return err
	}

	wb := new(bytes.Buffer)
	wb.Write(data)

	slice := strings.Split(requestBody.Extension, "/")

	bucketName := "my-blog-storage"
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(createdId + "." + slice[1]),
		Body:        wb,
		ContentType: aws.String(requestBody.Extension),
		ACL:         aws.String("public-read"),
	})

	if err != nil {
		return err
	}

	return nil
}

func main() {
	lambda.Start(Handler)
}
