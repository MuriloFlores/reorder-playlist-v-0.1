package entities

import (
	"sort"
	"time"
)

type playlist struct {
	id          string
	channelId   string
	title       string
	description string
	publishedAt time.Time
	videos      []VideoInterface
}

type PlaylistInterface interface {
	Id() string
	ChannelId() string
	Title() string
	Description() string
	PublishedAt() time.Time
	Videos() []VideoInterface
	SortByPublishedAt()
	SortByTitle()
	SortByDuration()
}

func NewPlaylist(id, channelId, title, description string, publishedAt time.Time, videos []VideoInterface) PlaylistInterface {
	return &playlist{
		id:          id,
		channelId:   channelId,
		title:       title,
		description: description,
		publishedAt: publishedAt,
		videos:      videos,
	}
}

func (p *playlist) Id() string {
	return p.id
}

func (p *playlist) ChannelId() string {
	return p.channelId
}

func (p *playlist) Title() string {
	return p.title
}

func (p *playlist) Description() string {
	return p.description
}

func (p *playlist) PublishedAt() time.Time {
	return p.publishedAt
}

func (p *playlist) Videos() []VideoInterface {
	return p.videos
}

func (p *playlist) SortByPublishedAt() {
	sort.Slice(p.videos, func(i, j int) bool {
		return p.videos[i].PublishedAt().Before(p.videos[j].PublishedAt())
	})
}

func (p *playlist) SortByTitle() {
	sort.Slice(p.videos, func(i, j int) bool {
		return p.videos[i].Title() < p.videos[j].Title()
	})
}

func (p *playlist) SortByDuration() {
	sort.Slice(p.videos, func(i, j int) bool {
		return p.videos[i].PublishedAt().After(p.videos[j].PublishedAt())
	})
}
