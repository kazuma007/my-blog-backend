package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	S3AK string
	S3SK string
}

type RequestBody struct {
	Tag string `json:"tag"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestBody := new(RequestBody)
	if err := json.Unmarshal(([]byte)(request.Body), requestBody); err != nil {
		fmt.Println("json unmarshal error", err)
	}

	var env Env
	if err := envconfig.Process("", &env); err != nil {
		fmt.Println("envconfig error", err)
		return events.APIGatewayProxyResponse{}, nil
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(env.S3AK, env.S3SK, ""),
		Region:      aws.String("ap-northeast-1"),
	}))

	if err := updateDynamoDB(requestBody, sess); err != nil {
		fmt.Println("updateDynamoDB error", err)
		return events.APIGatewayProxyResponse{}, nil
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

func updateDynamoDB(requestBody *RequestBody, sess *session.Session) error {
	svc := dynamodb.New(sess)

	time.Local = time.FixedZone("Asia/Tokyo", 9*60*60)
	time.LoadLocation("Asia/Tokyo")
	t := time.Now()
	layout := "2006/01/02T15:04:05"
	registeredTime := t.Format(layout)

	putParams := &dynamodb.PutItemInput{
		TableName: aws.String("my-blog-tag-t"),
		Item: map[string]*dynamodb.AttributeValue{
			"tag": {
				S: aws.String(requestBody.Tag),
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

func main() {
	lambda.Start(Handler)
}
