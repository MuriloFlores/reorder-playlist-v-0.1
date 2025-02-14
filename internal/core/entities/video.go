package entities

import "time"

type video struct {
	id          string
	title       string
	artist      string
	publishedAt time.Time
	duration    time.Duration
}

type VideoInterface interface {
	Id() string
	Title() string
	Artist() string
	PublishedAt() time.Time
	Duration() time.Duration
}

func NewVideo(id, title, artist string, publishedAt time.Time, duration time.Duration) VideoInterface {
	return &video{
		id:          id,
		title:       title,
		artist:      artist,
		publishedAt: publishedAt,
		duration:    duration,
	}
}

func (v *video) Id() string {
	return v.id
}

func (v *video) Title() string {
	return v.title
}

func (v *video) Artist() string {
	return v.artist
}

func (v *video) PublishedAt() time.Time {
	return v.publishedAt
}

func (v *video) Duration() time.Duration {
	return v.duration
}
