package repository_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"project/internal/DTOs"
	"project/internal/core/entities"
	"project/internal/infrastructure/repository"
)

type fakeCache struct {
	data map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{
		data: make(map[string]string),
	}
}

func (f *fakeCache) Set(key string, value interface{}, expiration time.Duration) error {
	switch v := value.(type) {
	case string:
		f.data[key] = v
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		f.data[key] = string(bytes)
	}
	return nil
}

func (f *fakeCache) Get(key string) (interface{}, error) {
	val, ok := f.data[key]
	if !ok {
		return nil, errors.New("redis: nil")
	}
	return val, nil
}

func (f *fakeCache) Delete(key string) error {
	delete(f.data, key)
	return nil
}

func TestSaveAndGetAllPlaylistsByUserID(t *testing.T) {
	fc := newFakeCache()
	repo := repository.NewPlaylistRepositoryRedis(fc)

	userID := "user123"

	playlist1 := entities.NewPlaylist("playlist1", "channel1", "Playlist 1", "Description 1", time.Now(), nil)
	playlist2 := entities.NewPlaylist("playlist2", "channel1", "Playlist 2", "Description 2", time.Now(), nil)
	playlists := []entities.PlaylistInterface{playlist1, playlist2}

	err := repo.SaveAllPlaylists(userID, playlists)
	if err != nil {
		t.Fatalf("Erro ao salvar playlists: %v", err)
	}

	retrieved, err := repo.GetAllPlaylistsByUserID(userID)
	if err != nil {
		t.Fatalf("Erro ao recuperar playlists: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("Esperado 2 playlists, obtido %d", len(retrieved))
	}

	if retrieved[0].Title() != playlist1.Title() && retrieved[1].Title() != playlist2.Title() {
		t.Errorf("Dados das playlists não correspondem ao esperado")
	}
}

func TestSaveGetAndDeletePlaylist(t *testing.T) {
	fc := newFakeCache()
	repo := repository.NewPlaylistRepositoryRedis(fc)

	userID := "user123"
	playlist := entities.NewPlaylist("playlist1", "channel1", "Playlist 1", "Description 1", time.Now(), nil)

	err := repo.SavePlaylist(userID, playlist)
	if err != nil {
		t.Fatalf("Erro ao salvar a playlist: %v", err)
	}

	key := "playlist1"
	data, err := json.Marshal(DTOs.PlaylistFromEntity(playlist))
	if err != nil {
		t.Fatalf("Erro ao serializar playlist: %v", err)
	}
	fc.data[key] = string(data)

	retrieved, err := repo.GetPlaylistById("playlist1")
	if err != nil {
		t.Fatalf("Erro ao recuperar a playlist: %v", err)
	}

	if retrieved.Title() != playlist.Title() {
		t.Errorf("Esperado título '%s', obtido '%s'", playlist.Title(), retrieved.Title())
	}

	err = repo.DeletePlaylist("playlist1")
	if err != nil {
		t.Fatalf("Erro ao deletar a playlist: %v", err)
	}

	_, err = repo.GetPlaylistById("playlist1")
	if err == nil {
		t.Errorf("Esperado erro ao buscar playlist deletada, mas não ocorreu")
	}
}
