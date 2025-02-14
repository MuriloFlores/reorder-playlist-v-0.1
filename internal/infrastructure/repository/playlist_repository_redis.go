package repository

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"strings"

	"github.com/redis/go-redis/v9"
	"project/internal/DTOs"
	"project/internal/core/entities"
	"project/internal/infrastructure/cache"
	"project/internal/infrastructure/logging"
)

type playlistRepositoryRedis struct {
	client cache.RedisCacheInterface
}

type PlaylistRepositoryRedisInterface interface {
	GetPlaylistById(playlistId string, userId string) (entities.PlaylistInterface, error)
	SavePlaylist(userId string, playlist entities.PlaylistInterface) error
	SaveAllPlaylists(userId string, playlists []entities.PlaylistInterface) error
	DeletePlaylist(playlistId string) error
	GetAllPlaylistsByUserID(userId string) ([]entities.PlaylistInterface, error)
}

func NewPlaylistRepositoryRedis(client cache.RedisCacheInterface) PlaylistRepositoryRedisInterface {
	return &playlistRepositoryRedis{client: client}
}

func (pr *playlistRepositoryRedis) GetPlaylistById(playlistId, userId string) (entities.PlaylistInterface, error) {
	key := "playlists:" + userId
	var dto DTOs.PlaylistRedisDTO

	data, err := pr.client.HGet(key, playlistId)
	if err != nil {
		logging.Error("GetPlaylistById - playlist_repository_redis - ln37", zap.String("Error", err.Error()))
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("playlist not found")
		}
		return nil, err
	}

	if err = json.Unmarshal([]byte(data.(string)), &dto); err != nil {
		return nil, err
	}
	return dto.ToEntity(), nil
}

func (pr *playlistRepositoryRedis) SavePlaylist(userId string, playlist entities.PlaylistInterface) error {
	key := "playlists:" + userId
	data, err := json.Marshal(playlist)
	if err != nil {
		return err
	}

	if err := pr.client.HSet(key, playlist.Id(), string(data)); err != nil {
		return err
	}

	return pr.client.HSet("playlistIndex", playlist.Id(), userId)
}

func (pr *playlistRepositoryRedis) SaveAllPlaylists(userId string, playlists []entities.PlaylistInterface) error {
	for _, playlist := range playlists {
		if err := pr.SavePlaylist(userId, playlist); err != nil {
			return err
		}
	}
	return nil
}

func (pr *playlistRepositoryRedis) DeletePlaylist(playlistId string) error {
	userId, err := pr.client.HGet("playlistIndex", playlistId)
	if err != nil {
		return err
	}
	if userId == "" {
		return errors.New("playlist not found in global index")
	}

	key := "playlists:" + userId.(string)
	if err := pr.client.HDel(key, playlistId); err != nil {
		return err
	}

	return pr.client.HDel("playlistIndex", playlistId)
}

func (pr *playlistRepositoryRedis) GetAllPlaylistsByUserID(userId string) ([]entities.PlaylistInterface, error) {
	key := "playlists:" + userId
	data, err := pr.client.HGetAll(key)
	if err != nil {
		return nil, err
	}

	var playlists []entities.PlaylistInterface
	for _, value := range data {
		var dto DTOs.PlaylistRedisDTO
		if err := json.Unmarshal([]byte(value), &dto); err != nil {
			logging.Error("Erro ao deserializar playlist", zap.Error(err))
			continue
		}
		if strings.TrimSpace(dto.Id) == "" || strings.TrimSpace(dto.Title) == "" {
			continue
		}
		playlists = append(playlists, dto.ToEntity())
	}

	return playlists, nil
}
