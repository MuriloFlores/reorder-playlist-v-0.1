package errors

type YouTubeErrorHandler interface {
	HandleYouTubeError(err error, playlistId, action string) error
}
