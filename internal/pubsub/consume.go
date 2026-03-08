package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	QueueTypeDurable   SimpleQueueType = "durable"
	QueueTypeTransient SimpleQueueType = "transient"
)

type AckType string

const (
	Ack         AckType = "Ack"
	NackRequeue AckType = "Nack and requeue"
	NackDiscard AckType = "Nack and discard"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
	handler func(T) AckType,
) error {
	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return fmt.Errorf("could not subscribe to %s: %v", queueName, err)
	}

	msgs, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	unmarshaller := func(data []byte) (T, error) {
		out := new(T)
		err := json.Unmarshal(data, &out)
		return *out, err
	}

	go func() {
		for msg := range msgs {
			out, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("error unmarshalling message: %v\n", err)
			}
			ackType := handler(out)
			switch ackType {
			case Ack:
				msg.Ack(false)
				fmt.Println("Message acked!")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Println("Message nacked and requeued!")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Println("Message nacked and discarded!")
			}
		}
	}()

	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not create channel: %v", err)
	}

	var durable bool
	var autoDelete bool
	var exclusive bool
	if queueType == QueueTypeDurable {
		durable = true
		autoDelete = false
		exclusive = false
	} else {
		durable = false
		autoDelete = true
		exclusive = true
	}

	args := amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	queue, err := channel.QueueDeclare(queueName, durable, autoDelete, exclusive, false, args)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	err = channel.QueueBind(queue.Name, key, exchange, false, nil)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}

	return channel, queue, nil
}
