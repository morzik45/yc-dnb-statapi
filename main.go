package main

import (
	"context"
	"encoding/json"
	"fmt"
)

// Структура запроса (см. https://cloud.yandex.ru/docs/functions/concepts/function-invoke#request)
// Остальные поля нигде не используются в данном примере, поэтому можно обойтись без них
type RequestBody struct {
	HttpMethod string `json:"httpMethod"`
	Body       []byte `json:"body"`
}

// Преобразуем поле body объекта RequestBody
type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int         `json:"statusCode"`
	Body       interface{} `json:"body"`
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

	name := req.Name
	// Тело ответа необходимо вернуть в виде структуры, которая автоматически преобразуется в JSON-документ,
	// который отобразится на экране
	return &Response{
		StatusCode: 200,
		Body:       fmt.Sprintf("Hello, %s", name),
	}, nil
}