package messaging

import (
	"fmt"

	"github.com/streadway/amqp"
)

// RabbitMQProducerInterface define os métodos para publicar mensagens.
type RabbitMQProducerInterface interface {
	Publish(message string) error
	Close()
}

type rabbitmqProducer struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue amqp.Queue
}

func NewRabbitMQProducer(queueName string) RabbitMQProducerInterface {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(fmt.Sprintf("Falha ao conectar ao RabbitMQ: %v", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Falha ao abrir canal: %v", err))
	}

	q, err := ch.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		panic(fmt.Sprintf("Falha ao declarar fila: %v", err))
	}

	return &rabbitmqProducer{conn: conn, ch: ch, queue: q}
}

func (p *rabbitmqProducer) Publish(message string) error {
	err := p.ch.Publish(
		"", // exchange vazio para envio direto à fila
		p.queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	)
	if err != nil {
		return fmt.Errorf("falha ao publicar mensagem: %v", err)
	}
	return nil
}

func (p *rabbitmqProducer) Close() {
	p.ch.Close()
	p.conn.Close()
}
