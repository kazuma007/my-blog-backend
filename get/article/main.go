package main

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/kelseyhightower/envconfig"
)

type Env struct {
	S3AK string
	S3SK string
}

type ResponseData struct {
	Key            string `json:"storage_key", dynamodbav:"storage_key"`
	Title          string `json:"title", dynamodbav:"title"`
	Content        string `json:"content", dynamodbav:"content"`
	RegisteredTime string `json:"registered_time", dynamodbav:"registered_time"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var env Env
	if err := envconfig.Process("", &env); err != nil {
		fmt.Println("envconfig error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(env.S3AK, env.S3SK, ""),
		Region:      aws.String("ap-northeast-1"),
	}))

	svc := dynamodb.New(sess)

	key := request.QueryStringParameters["key"]

	params := &dynamodb.GetItemInput{
		TableName: aws.String("my-blog-t"),
		Key: map[string]*dynamodb.AttributeValue{
			"storage_key": {
				S: aws.String(key),
			},
		},
	}

	result, err := svc.GetItem(params)
	if err != nil {
		fmt.Println("dynamoDB scan error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	var responseData ResponseData
	dynamodbattribute.UnmarshalMap(result.Item, &responseData)

	jsonResponse, err := json.Marshal(responseData)
	if err != nil {
		fmt.Println("json marshal error", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "*",
			"Access-Control-Allow-Headers": "origin,Accept,Authorization,Content-Type",
			"Content-Type":                 "application/json",
		},
		Body:            string(jsonResponse),
		IsBase64Encoded: false,
		StatusCode:      200,
	}, nil
}

func main() {
	lambda.Start(Handler)
}
