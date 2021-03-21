package main

import (
	"context"
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

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

type Env struct {
	S3AK string
	S3SK string
}

type ResponseData struct {
	Key            string `json:"storage_key", dynamodbav:"storage_key"`
	Filename       string `json:"filename", dynamodbav:"filename"`
	Title          string `json:"title", dynamodbav:"title"`
	Content        string `json:"content", dynamodbav:"content"`
	RegisteredTime string `json:"registered_time", dynamodbav:"registered_time"`
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
func Handler(ctx context.Context) (Response, error) {
	var env Env
	if err := envconfig.Process("", &env); err != nil {
		fmt.Println("envconfig error", err)
		return Response{StatusCode: 500}, err
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(env.S3AK, env.S3SK, ""),
		Region:      aws.String("ap-northeast-1"),
	}))

	svc := dynamodb.New(sess)

	params := &dynamodb.ScanInput{
		TableName: aws.String("my-blog-t"),
		Limit:     aws.Int64(50),
	}

	result, err := svc.Scan(params)
	if err != nil {
		fmt.Println("dynamoDB scan error", err)
		return Response{StatusCode: 500}, err
	}

	var responseData []ResponseData
	dynamodbattribute.UnmarshalListOfMaps(result.Items, &responseData)

	jsonResponse, err := json.Marshal(responseData)
	if err != nil {
		fmt.Println("json marshal error", err)
		return Response{StatusCode: 500}, err
	}

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            string(jsonResponse),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(Handler)
}
