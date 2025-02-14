package entities

import "time"

type video struct {
	id          string
	title       string
	channelId   string
	language    string
	publishedAt time.Time
	duration    time.Duration
}

type VideoInterface interface {
	Id() string
	Title() string
	ChannelId() string
	Language() string
	PublishedAt() time.Time
	Duration() time.Duration
}

func NewVideo(id, title, channelId, language string, publishedAt time.Time, duration time.Duration) VideoInterface {
	return &video{
		id:          id,
		title:       title,
		channelId:   channelId,
		language:    language,
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

func (v *video) ChannelId() string {
	return v.channelId
}

func (v *video) Language() string {
	return v.language
}

func (v *video) PublishedAt() time.Time {
	return v.publishedAt
}

func (v *video) Duration() time.Duration {
	return v.duration
}
