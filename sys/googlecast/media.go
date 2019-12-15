/*
	Go Language Raspberry Pi Interface
	(c) Copyright David Thorpe 2019
	All Rights Reserved
	Documentation http://djthorpe.github.io/gopi/
	For Licensing and Usage information, please see LICENSE.md
*/

package googlecast

import (
	"fmt"
	"strconv"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

type media struct {
	MediaSessionId int       `json:"mediaSessionId"`
	PlayerState    string    `json:"playerState"`
	CurrentTime    float32   `json:"currentTime"`
	IdleReason     string    `json:"idleReason"`
	Volume         volume    `json:"volume"`
	CurrentItemId  int       `json:"currentItemId"`
	LoadingItemId  int       `json:"loadingItemId"`
	Media          mediaItem `json:"media"`
}

type mediaItem struct {
	ContentId   string        `json:"contentId"`
	ContentType string        `json:"contentType"`
	StreamType  string        `json:"streamType"`
	Duration    float32       `json:"duration"`
	Metadata    mediaMetadata `json:"metadata"`
}

type mediaMetadata struct {
	MetadataType int          `json:"metadataType`
	Artist       string       `json:"artist"`
	Title        string       `json:"title"`
	Subtitle     string       `json:"subtitle"`
	Images       []mediaImage `json:"images"`
	ReleaseDate  string       `json:"releaseDate"`
}

type mediaImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

////////////////////////////////////////////////////////////////////////////////
// IMPLEMENTATION

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (this *media) String() string {
	var parts string
	if this.PlayerState != "" {
		parts += fmt.Sprintf(" state=%v", strconv.Quote(this.PlayerState))
	}
	if this.IdleReason != "" {
		parts += fmt.Sprintf(" idle_reason=%v", strconv.Quote(this.IdleReason))
	}
	if this.CurrentTime != 0 {
		parts += fmt.Sprintf(" current_time=%v", this.CurrentTime)
	}
	if this.CurrentItemId != 0 {
		parts += fmt.Sprintf(" current_id=%v", this.CurrentItemId)
	}
	if this.LoadingItemId != 0 {
		parts += fmt.Sprintf(" loading_id=%v", this.LoadingItemId)
	}
	if this.Media.ContentId != "" {
		parts += fmt.Sprintf(" %v", this.Media)
	}
	return fmt.Sprintf("<media>{ id=%v%v }", this.MediaSessionId, parts)
}

func (this mediaItem) String() string {
	var parts string
	if this.ContentType != "" {
		parts += fmt.Sprintf(" content_type=%v", strconv.Quote(this.ContentType))
	}
	if this.StreamType != "" {
		parts += fmt.Sprintf(" stream_type=%v", strconv.Quote(this.StreamType))
	}
	if this.Duration != 0 {
		parts += fmt.Sprintf(" duration=%v", this.Duration)
	}
	if this.Metadata.MetadataType != 0 {
		parts += fmt.Sprintf(" %v", this.Metadata)
	}
	return fmt.Sprintf("<mediaItem>{ id=%v%v }", this.ContentId, parts)
}

func (this mediaMetadata) String() string {
	var parts string
	if this.Artist != "" {
		parts += fmt.Sprintf(" artist=%v", strconv.Quote(this.Artist))
	}
	if this.Title != "" {
		parts += fmt.Sprintf(" title=%v", strconv.Quote(this.Title))
	}
	if this.Subtitle != "" {
		parts += fmt.Sprintf(" subtitle=%v", strconv.Quote(this.Subtitle))
	}
	if this.ReleaseDate != "" {
		parts += fmt.Sprintf(" release_date=%v", strconv.Quote(this.ReleaseDate))
	}
	if len(this.Images) > 0 {
		parts += fmt.Sprintf(" images=%v", this.Images)
	}
	return fmt.Sprintf("<mediaMetadata>{ type=%v%v }", this.MetadataType, parts)
}

func (this mediaImage) String() string {
	var parts string
	if this.Width != 0 {
		parts += fmt.Sprintf(" w=%v", this.Width)
	}
	if this.Height != 0 {
		parts += fmt.Sprintf(" h=%v", this.Height)
	}
	return fmt.Sprintf("<mediaMetadata>{ url=%v%v }", this.URL, parts)
}
