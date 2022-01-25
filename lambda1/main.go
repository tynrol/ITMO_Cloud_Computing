package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	REGION      = ""
	BUCKET_NAME = ""
	QUEUE       = ""
	BOT_TOKEN   = ""
)

var s3session *s3.S3

type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func init() {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(REGION),
	}))

	s3session = s3.New(sess)
}

func main() {
	lambda.Start(Handler)
}

func Handler(sqsEvent events.SQSEvent) {
	for _, message := range sqsEvent.Records {
		fmt.Printf("The message %s for event source %s = %s \n", message.MessageId, message.EventSource, message.Body)
		Encrypt(message)
	}
}

func Encrypt(message events.SQSMessage) {
	s := strings.Split(message.Body, "_")
	charID, err := strconv.Atoi(s[0])
	if err != nil {
		fmt.Println("Cant parse a number")
	}

	t := time.Now().Format("01-02-2006_15-04-05")
	filename := fmt.Sprintf("%s_%s.img", s[0], t)
	fmt.Println(filename)
	png, err := qrcode.Encode(s[1], qrcode.Medium, 256)
	fmt.Println(png)
	if err != nil {
		fmt.Println("Can't create qrcode")
	}

	// _, err = s3session.PutObject(&s3.PutObjectInput{
	// 	Bucket: aws.String(BUCKET_NAME),
	// 	Key:    aws.String(filename),
	// 	Body:   bytes.NewReader(png),
	// })
	// if err != nil {
	// 	fmt.Println("Cant put obj on s3")
	// }

	_, err = s3session.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(BUCKET_NAME),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(png),
	})
	if err != nil {
		// HandleError(input.Message.Chat.ID, "Couldnt place file into bucket")
		// return events.APIGatewayProxyResponse{Body: request.Body, StatusCode: 411}, err
		fmt.Println("Cant put obj on s3")
	}

	//322295846
	Send(int64(charID), filename)
}

func Send(chatID int64, link string) {

	fmt.Println("Entered Handle error")
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   getS3Link(link),
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

func getS3Link(key string) string {
	return fmt.Sprintf("https://%s.%s.amazonaws.com/%s", BUCKET_NAME, REGION, key)
}
