package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	jsonData, err := json.Marshal(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	encodedData, err := encode(val)
	if err != nil {
		return err
	}

	msg := amqp.Publishing{
		ContentType: "application/gob",
		Body:        encodedData,
	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
}

func encode[T any](val T) ([]byte, error) {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(&val)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func decode[T any](data []byte) (T, error) {
	var val T

	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)

	err := dec.Decode(&val)
	if err != nil {
		return val, err
	}

	return val, nil
}
