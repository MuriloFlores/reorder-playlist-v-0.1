package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strings"

	"github.com/streadway/amqp"
	"project/internal/DTOs"
	coreErrors "project/internal/core/errors"
	"project/internal/core/services"
	"project/internal/infrastructure/logging"
)

// RabbitMQConsumer consome mensagens da fila e processa ações.
type RabbitMQConsumer struct {
	Service      services.YoutubePlaylistService
	ErrorHandler coreErrors.YouTubeErrorHandler
}

func NewRabbitMQConsumer(service services.YoutubePlaylistService, errorHandler coreErrors.YouTubeErrorHandler) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		Service:      service,
		ErrorHandler: errorHandler,
	}
}

func (c *RabbitMQConsumer) StartRabbitMQConsumer(queueName string) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(fmt.Sprintf("Falha ao conectar RabbitMQ: %v", err))
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("Falha ao abrir canal: %v", err))
	}
	defer ch.Close()

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		panic(fmt.Sprintf("Falha ao registrar consumidor: %v", err))
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			logging.Info(fmt.Sprintf("Mensagem recebida: %s", d.Body))
			var action DTOs.PlaylistActionDTO
			if err := json.Unmarshal(d.Body, &action); err != nil {
				logging.Error("Erro ao decodificar mensagem", zap.String("error: ", err.Error()))
				d.Nack(false, false)
				continue
			}
			parts := strings.Split(action.ActionName, "_")
			if len(parts) > 0 && parts[0] == "reorder" {
				err := c.Service.ReorderPlaylist(action.PlaylistId, action.Params, action.UserId, context.Background())
				if err != nil {
					logging.Error("Erro ao reordenar playlist", zap.String("error: ", err.Error()))
					d.Nack(false, true)
					continue
				}
			}
			d.Ack(false)
		}
	}()
	logging.Info("Consumidor está em execução...")
	<-forever
}
