package services

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"project/internal/infrastructure/logging"
	"project/internal/infrastructure/sessions"
	"time"

	"github.com/sosodev/duration"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"project/internal/core/entities"
	coreErrors "project/internal/core/errors"
	"project/internal/infrastructure/repository"
)

// YoutubePlaylistService define as operações para gerenciar playlists do YouTube.
type YoutubePlaylistService interface {
	GetAllPlaylists(ctx context.Context, token *oauth2.Token, r *http.Request) ([]entities.PlaylistInterface, error)
	ReorderPlaylist(playlistId, criteria, userId string, ctx context.Context) error
	DeletePlaylist(playlistId string) error
	GetPlaylistVideos(playlistId string) ([]entities.VideoInterface, error)
	GetVideoDetails(videoId string) (entities.VideoInterface, error)
	CreateNewPlaylist(playlist entities.PlaylistInterface) (string, error)
}

type youtubePlaylistService struct {
	repo         repository.PlaylistRepositoryRedisInterface
	Youtube      *youtube.Service
	errorHandler coreErrors.YouTubeErrorHandler
	session      sessions.SessionManager
}

func NewYoutubePlaylistService(repo repository.PlaylistRepositoryRedisInterface, eh coreErrors.YouTubeErrorHandler, session sessions.SessionManager) YoutubePlaylistService {
	return &youtubePlaylistService{
		repo:         repo,
		errorHandler: eh,
		session:      session,
	}
}

func (s *youtubePlaylistService) getYoutubeService(ctx context.Context, token *oauth2.Token) (*youtube.Service, error) {
	if s.Youtube == nil {
		service, err := youtube.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
		if err != nil {
			return nil, err
		}
		s.Youtube = service
	}
	return s.Youtube, nil
}

func (s *youtubePlaylistService) GetAllPlaylists(ctx context.Context, token *oauth2.Token, r *http.Request) ([]entities.PlaylistInterface, error) {
	// Obter serviço do YouTube
	ytService, err := s.getYoutubeService(ctx, token)
	if err != nil {
		logging.Error("Erro ao obter serviço do YouTube", zap.Error(err))
		return nil, err
	}

	// Extrair userId da sessão
	userId := s.session.GetUserId(r)
	logging.Info("UserID extraído da sessão", zap.String("userID", userId))

	// Tentar recuperar playlists do cache (Redis)
	playlistsByUserID, err := s.repo.GetAllPlaylistsByUserID(userId)
	if err == nil && len(playlistsByUserID) > 0 {
		logging.Info("Playlists recuperadas do cache", zap.Int("count", len(playlistsByUserID)))

		return playlistsByUserID, nil
	}

	for i, playlist := range playlistsByUserID {
		log.Printf("Playlist %d: %+v\n", i, playlist)
	}

	// Chamar API do YouTube para obter playlists
	call := ytService.Playlists.List([]string{"id", "snippet", "contentDetails"}).Mine(true).MaxResults(50)
	response, err := call.Do()
	if err != nil {
		logging.Error("Erro ao chamar API do YouTube", zap.Error(err))
		return nil, s.errorHandler.HandleYouTubeError(err, "", "get_all_playlists")
	}

	logging.Info("Número de playlists retornadas pela API", zap.Int("count", len(response.Items)))
	if len(response.Items) == 0 {
		logging.Info("Nenhuma playlist encontrada na API")
		return []entities.PlaylistInterface{}, nil
	}

	// Converter a resposta da API em entidades do domínio
	playlistsEntity := make([]entities.PlaylistInterface, len(response.Items))
	for i, item := range response.Items {
		logging.Info("Processando playlist", zap.String("playlistID", item.Id), zap.String("title", item.Snippet.Title))
		publishTime, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil {
			logging.Error("Erro ao converter data de publicação", zap.Error(err), zap.String("publishedAt", item.Snippet.PublishedAt))
			return nil, err
		}

		videos, err := s.GetPlaylistVideos(item.Id)
		if err != nil {
			logging.Info("error getting videos")
			return nil, err
		}

		playlistsEntity[i] = entities.NewPlaylist(
			item.Id,
			item.Snippet.ChannelId,
			item.Snippet.Title,
			item.Snippet.Description,
			publishTime,
			videos,
		)
	}

	logging.Info("Salvando playlists no cache", zap.String("userID", userId), zap.Int("count", len(playlistsEntity)))
	err = s.repo.SaveAllPlaylists(userId, playlistsEntity)
	if err != nil {
		logging.Error("Erro ao salvar playlists no cache", zap.Error(err))
		return nil, s.errorHandler.HandleYouTubeError(err, "", "save_all_playlists")
	}
	logging.Info("Playlists salvas no cache com sucesso", zap.Int("count", len(playlistsEntity)))

	return playlistsEntity, nil
}

func (s *youtubePlaylistService) ReorderPlaylist(playlistId, criteria, userId string, ctx context.Context) error {
	ytService := s.Youtube

	playlist, err := s.GetPlaylistByID(ctx, ytService, playlistId)
	if err != nil {
		logging.Info("Error getting playlist")
		return s.errorHandler.HandleYouTubeError(err, playlistId, "reorder_playlist")
	}

	switch criteria {
	case "byTitle":
		playlist.SortByTitle()
	case "byPublishedAt":
		playlist.SortByPublishedAt()
	case "byDuration":
		playlist.SortByDuration()
	default:
		return errors.New("invalid criteria")
	}

	fmt.Println("--------------------------")
	fmt.Println("Reorder playlist:", playlist)
	fmt.Println("--------------------------")

	_, err = s.CreateNewPlaylist(playlist)
	if err != nil {
		logging.Info("Erro creating a new playlist - youtube_service - ln 153", zap.Error(err))
		return s.errorHandler.HandleYouTubeError(err, playlistId, "reorder_playlist")
	}

	return s.repo.SavePlaylist(userId, playlist)
}

func (s *youtubePlaylistService) DeletePlaylist(playlistId string) error {
	err := s.Youtube.Playlists.Delete(playlistId).Do()
	if err != nil {
		return s.errorHandler.HandleYouTubeError(err, playlistId, "delete_playlist")
	}
	return s.repo.DeletePlaylist(playlistId)
}

func (s *youtubePlaylistService) GetPlaylistVideos(playlistId string) ([]entities.VideoInterface, error) {
	var videos []entities.VideoInterface
	pageToken := ""
	for {
		videoIds, nextPageToken, err := s.getPlaylistVideoIds(playlistId, pageToken)
		if err != nil {
			return nil, s.errorHandler.HandleYouTubeError(err, playlistId, "get_playlist_videos")
		}

		for _, videoId := range videoIds {
			video, err := s.GetVideoDetails(videoId)
			if err != nil {
				log.Printf("Erro ao buscar detalhes do vídeo %s: %v", videoId, err)
				continue
			}
			videos = append(videos, video)
		}

		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}

	return videos, nil
}

func (s *youtubePlaylistService) GetVideoDetails(videoId string) (entities.VideoInterface, error) {
	call := s.Youtube.Videos.List([]string{"snippet", "contentDetails"}).Id(videoId)
	response, err := call.Do()
	if err != nil {
		return nil, s.errorHandler.HandleYouTubeError(err, videoId, "get_video_details")
	}
	if len(response.Items) == 0 {
		return nil, errors.New("video not found")
	}
	item := response.Items[0]
	parsedDuration, err := duration.Parse(item.ContentDetails.Duration)
	if err != nil {
		return nil, err
	}
	publishedAt, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		return nil, err
	}
	video := entities.NewVideo(item.Id, item.Snippet.Title, item.Snippet.ChannelId, publishedAt, parsedDuration.ToTimeDuration())
	return video, nil
}

func (s *youtubePlaylistService) getPlaylistVideoIds(playlistId, pageToken string) ([]string, string, error) {
	call := s.Youtube.PlaylistItems.List([]string{"contentDetails"}).PlaylistId(playlistId).MaxResults(50).PageToken(pageToken)
	response, err := call.Do()
	if err != nil {
		return nil, "", s.errorHandler.HandleYouTubeError(err, playlistId, "get_playlist_video_ids")
	}

	var ids []string
	for _, item := range response.Items {
		if item.ContentDetails == nil || item.ContentDetails.VideoId == "" {
			continue
		}
		ids = append(ids, item.ContentDetails.VideoId)
	}
	return ids, response.NextPageToken, nil
}

func (s *youtubePlaylistService) CreateNewPlaylist(playlist entities.PlaylistInterface) (string, error) {
	call := s.Youtube.Playlists.Insert([]string{"snippet", "status"}, &youtube.Playlist{
		Snippet: &youtube.PlaylistSnippet{
			Title:       playlist.Title() + " reorder_playlist_" + time.Now().Format(time.RFC3339),
			Description: playlist.Description(),
		},
		Status: &youtube.PlaylistStatus{
			PrivacyStatus: "public",
		},
	})
	response, err := call.Do()
	if err != nil {
		logging.Error("Error on line 251")
		return "", s.errorHandler.HandleYouTubeError(err, playlist.Id(), "create_playlist")
	}

	for _, video := range playlist.Videos() {
		fmt.Println("-------------------------------------")
		fmt.Println(video)
		fmt.Println("--------------------------")
		err := s.addVideoToPlaylist(response.Id, video.Id())
		if err != nil {
			logging.Error("Erro ao adicionar video a nova playlist", zap.String("video_id", video.Id()), zap.Error(err))
			continue
		}
	}
	return response.Id, nil
}

func (s *youtubePlaylistService) addVideoToPlaylist(playlistId, videoId string) error {
	call := s.Youtube.PlaylistItems.Insert([]string{"snippet"}, &youtube.PlaylistItem{
		Snippet: &youtube.PlaylistItemSnippet{
			PlaylistId: playlistId,
			ResourceId: &youtube.ResourceId{
				Kind:    "youtube#video",
				VideoId: videoId,
			},
		},
	})
	_, err := call.Do()
	if err != nil {
		return s.errorHandler.HandleYouTubeError(err, playlistId, "add_video_to_playlist")
	}
	return nil
}

func (s *youtubePlaylistService) GetPlaylistByID(ctx context.Context, service *youtube.Service, playlistID string) (entities.PlaylistInterface, error) {
	// Cria a chamada especificando os parts que deseja recuperar.
	call := service.Playlists.List([]string{"snippet", "status", "contentDetails"}).Id(playlistID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar a playlist: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, errors.New("playlist não encontrada")
	}

	responseItem := response.Items[0]

	videos, err := s.GetPlaylistVideos(playlistID)
	if err != nil {
		logging.Error("Erro ao buscar os videos da playlist - youtube_service - ln 301")
		return nil, err
	}

	publishTime, err := time.Parse(time.RFC3339, responseItem.Snippet.PublishedAt)
	if err != nil {
		logging.Error("Erro ao converter data de publicação", zap.Error(err), zap.String("publishedAt", responseItem.Snippet.PublishedAt))
		return nil, err
	}

	playlist := entities.NewPlaylist(
		responseItem.Id,
		responseItem.Snippet.ChannelId,
		responseItem.Snippet.Title,
		responseItem.Snippet.Description,
		publishTime,
		videos,
	)

	return playlist, nil
}
