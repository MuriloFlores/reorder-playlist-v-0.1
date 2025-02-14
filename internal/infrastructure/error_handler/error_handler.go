package error_handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/api/googleapi"
	"project/internal/DTOs"
	coreErrors "project/internal/core/errors"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/messaging"
)

// ErrorHandler centraliza o tratamento de errors e integra com o sistema de filas.
type errorHandler struct {
	producer messaging.RabbitMQProducerInterface
}

func NewErrorHandler(producer messaging.RabbitMQProducerInterface) coreErrors.YouTubeErrorHandler {
	return &errorHandler{producer: producer}
}

// HandleQuotaExceededError trata o erro de cota excedida e agenda uma re-tentativa.
func (eh *errorHandler) HandleQuotaExceededError(playlistId string, err *googleapi.Error, action string) error {
	logging.Info(fmt.Sprintf("Quota excedida para playlist %s", playlistId))
	message := DTOs.PlaylistActionDTO{
		ActionName: action,
		PlaylistId: playlistId,
		Err:        err.Error(),
		RetryAt:    time.Now().Add(24 * time.Hour).Unix(),
	}
	jsonMessage, jErr := json.Marshal(message)
	if jErr != nil {
		return fmt.Errorf("erro ao serializar mensagem: %v", jErr)
	}
	if pubErr := eh.producer.Publish(string(jsonMessage)); pubErr != nil {
		return fmt.Errorf("erro ao publicar mensagem: %v", pubErr)
	}
	return fmt.Errorf("quota excedida. Ação agendada para reprocessamento em 24 horas")
}

// HandleYouTubeError centraliza o tratamento de errors da API do YouTube.
func (eh *errorHandler) HandleYouTubeError(err error, playlistId, action string) error {
	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		if gErr.Code == 403 && isQuotaExceeded(gErr) {
			return eh.HandleQuotaExceededError(playlistId, gErr, action)
		}
	}
	return fmt.Errorf("erro inesperado: %v", err)
}

func isQuotaExceeded(err *googleapi.Error) bool {
	for _, detail := range err.Errors {
		if detail.Reason == "quotaExceeded" {
			return true
		}
	}
	return false
}
