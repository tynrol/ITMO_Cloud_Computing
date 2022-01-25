package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Message struct {
	Message struct {
		MessageID int `json:"message_id"`
		From      struct {
			ID           int64  `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			ID        int64  `json:"id"`
			FirstName string `json:"first_name"`
			Username  string `json:"username"`
			Type      string `json:"type"`
		} `json:"chat"`
		Date int    `json:"date"`
		Text string `json:"text"`
	} `json:"message"`
}

type SQSMessage struct {
	Key    string `json:"key"`
	S3Link string `json:"s3link"`
	ChatId int64  `json:"id"`
}

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

var s3session *s3.S3
var sqsQueue *sqs.SQS

// var bot *tgbotapi.BotAPI

const (
	REGION      = ""
	BUCKET_NAME = ""
	QUEUE       = ""
	BOT_TOKEN   = ""
)

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	}))

	s3session = s3.New(sess)
	sqsQueue = sqs.New(sess)
}

func main() {
	lambda.Start(Handler)
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Processing request data for request %s.\n", request.RequestContext.RequestID)
	fmt.Printf("Body %s\n", request.Body)

	var input Message
	err := json.Unmarshal([]byte(request.Body), &input)
	if err != nil {
		HandleError(input.Message.Chat.ID, "Wrong request")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 410}, err
	}

	//filter this if needed
	// sqsMessage := &SQSMessage{
	// 	Key:    filename,
	// 	S3Link: fmt.Sprintf("https://%s.%s.amazonaws.com/%s", BUCKET_NAME, REGION, filename),
	// 	ChatId: input.Message.Chat.ID,
	// }

	chatIdStr := strconv.Itoa(int(input.Message.Chat.ID))

	_, err = sqsQueue.SendMessage(&sqs.SendMessageInput{
		// MessageAttributes: map[string]*sqs.MessageAttributeValue{
		// 	"Key": &sqs.MessageAttributeValue{
		// 		DataType:    aws.String("String"),
		// 		StringValue: aws.String(sqsMessage.Key),
		// 	},
		// 	"S3Link": &sqs.MessageAttributeValue{
		// 		DataType:    aws.String("String"),
		// 		StringValue: aws.String(sqsMessage.S3Link),
		// 	},
		// 	"ChatId": &sqs.MessageAttributeValue{
		// 		DataType:    aws.String("Number"),
		// 		StringValue: &sqsMessage.ChatId,
		// 	},
		// },
		MessageBody: aws.String(chatIdStr + "_" + input.Message.Text),
		QueueUrl:    aws.String(QUEUE),
	})
	if err != nil {
		HandleError(input.Message.Chat.ID, "Couldnt place message in SQS")
		return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 412}, err
	}
	return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 200}, nil
}

func HandleError(chatID int64, text string) {
	fmt.Println("Entered Handle error")
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Println(err)
	}

	res, err := http.Post("", "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		fmt.Println("Wasnt able to send", err)
	}
	if res.StatusCode != http.StatusOK {
		fmt.Println(res.Status)
	}
}
