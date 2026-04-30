package model

import "time"

type Playlist struct {
	ID                    int64      `json:"id"`
	Name                  string     `json:"name"`
	M3UURL                string     `json:"m3u_url"`
	EPGURL                string     `json:"epg_url"`
	UpdateIntervalMinutes int        `json:"update_interval_minutes"`
	Enabled               bool       `json:"enabled"`
	ChannelCount          int        `json:"channel_count"`
	EPGChannelCount       int        `json:"epg_channel_count"`
	LastSyncAt            *time.Time `json:"last_sync_at,omitempty"`
	LastError             string     `json:"last_error,omitempty"`
}

type Channel struct {
	ID               int64      `json:"id"`
	PlaylistID       int64      `json:"playlist_id"`
	Name             string     `json:"name"`
	Group            string     `json:"group"`
	SortIndex        int        `json:"sort_index"`
	Logo             string     `json:"logo"`
	StreamURL        string     `json:"stream_url"`
	ArchiveSupported bool       `json:"archive_supported"`
	ExternalID       string     `json:"external_id"`
	StreamModeCache  string     `json:"-"`
	StreamModeAt     *time.Time `json:"-"`
}

type Program struct {
	PlaylistID  int64     `json:"playlist_id"`
	ChannelID   int64     `json:"channel_id"`
	Title       string    `json:"title"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Description string    `json:"description"`
}
