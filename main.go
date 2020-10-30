package main

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

// Структура запроса (см. https://cloud.yandex.ru/docs/functions/concepts/function-invoke#request)
// Остальные поля нигде не используются в данном примере, поэтому можно обойтись без них
type RequestBody struct {
	HttpMethod string `json:"httpMethod"`
	Body       []byte `json:"body"`
}

// Преобразуем поле body объекта RequestBody
type Request struct {
	Name  string    `json:"name"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
}

type Data struct {
	Id    string `bson:"_id" json:"id"`
	Count int    `bson:"count" json:"count"`
}

func Handler(ctx context.Context, request []byte) (*Response, error) {
	requestBody := &RequestBody{}
	// Массив байтов, содержащий тело запроса, преобразуется в соответствующий объект
	err := json.Unmarshal(request, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when parsing request: %v", err)
	}

	// В журнале будет напечатано название HTTP-метода, с помощью которого осуществлен запрос, а так же тело запроса
	fmt.Println(requestBody.HttpMethod, string(requestBody.Body))

	req := &Request{}
	// Поле body запроса преобразуется в объект типа Request для получения переданного имени
	err = json.Unmarshal(requestBody.Body, &req)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred when parsing body: %v", err)
	}

	if req.Name == "new_users" {
		data, _ := GetData(*req)
		return &Response{
			StatusCode: 200,
			Body:       data,
		}, nil
	}
	// Тело ответа необходимо вернуть в виде структуры, которая автоматически преобразуется в JSON-документ,
	// который отобразится на экране
	return &Response{
		StatusCode: 200,
		Body:       "",
	}, nil
}

func GetData(req Request) ([]Data, error) {
	var ctx = context.TODO()
	var data []Data
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return data, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return data, err
	}
	Users := client.Database(os.Getenv("DB_NAME")).Collection("users")

	matchStage := bson.D{{"$match", bson.D{{"datetimes.first_visit", bson.D{{"$lt", time.Now()}, {"$gt", time.Now().Add(-time.Hour * 144)}}}}}}
	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"$dateToString", bson.D{{"format", "%Y-%m-%d"}, {"date", "$datetimes.first_visit"}}}}}, {"count", bson.D{{"$sum", 1}}}}}}
	sortStage := bson.D{{"$sort", bson.D{{"_id", 1}}}}

	result, err := Users.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, sortStage})
	err = result.All(ctx, &data)
	return data, err
}
