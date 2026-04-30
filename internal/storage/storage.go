package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"webtv/internal/model"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS playlists (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	m3u_url TEXT NOT NULL,
	epg_url TEXT NOT NULL,
	update_interval_minutes INTEGER NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	last_sync_at TEXT,
	last_error TEXT
);
CREATE TABLE IF NOT EXISTS channels (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	playlist_id INTEGER NOT NULL,
	external_id TEXT NOT NULL,
	name TEXT NOT NULL,
	group_name TEXT,
	sort_index INTEGER NOT NULL DEFAULT 0,
	logo TEXT,
	stream_url TEXT NOT NULL,
	archive_supported INTEGER NOT NULL DEFAULT 0,
	stream_mode_cache TEXT,
	stream_mode_checked_at TEXT,
	UNIQUE(playlist_id, external_id)
);
CREATE TABLE IF NOT EXISTS programs (
	playlist_id INTEGER NOT NULL DEFAULT 0,
	channel_id INTEGER NOT NULL,
	title TEXT NOT NULL,
	start_at TEXT NOT NULL,
	end_at TEXT NOT NULL,
	description TEXT,
	PRIMARY KEY(channel_id, start_at)
);
CREATE TABLE IF NOT EXISTS hidden_groups (
	playlist_id INTEGER NOT NULL,
	group_name TEXT NOT NULL,
	PRIMARY KEY(playlist_id, group_name)
);
CREATE TABLE IF NOT EXISTS hidden_channels (
	playlist_id INTEGER NOT NULL,
	channel_external_id TEXT NOT NULL,
	PRIMARY KEY(playlist_id, channel_external_id)
);`)
	if err != nil {
		return err
	}
	// Older DBs may not have playlist_id in programs; add it for isolation.
	if _, err := s.db.Exec(`ALTER TABLE programs ADD COLUMN playlist_id INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	if _, err := s.db.Exec(`ALTER TABLE channels ADD COLUMN sort_index INTEGER NOT NULL DEFAULT 0`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	if _, err := s.db.Exec(`ALTER TABLE channels ADD COLUMN stream_mode_cache TEXT`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	if _, err := s.db.Exec(`ALTER TABLE channels ADD COLUMN stream_mode_checked_at TEXT`); err != nil {
		if !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	return nil
}

func (s *Store) CreatePlaylist(ctx context.Context, p model.Playlist) (model.Playlist, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO playlists(name,m3u_url,epg_url,update_interval_minutes,enabled) VALUES(?,?,?,?,?)`, p.Name, p.M3UURL, p.EPGURL, p.UpdateIntervalMinutes, boolToInt(p.Enabled))
	if err != nil {
		return model.Playlist{}, err
	}
	id, _ := res.LastInsertId()
	p.ID = id
	return p, nil
}

func (s *Store) ListPlaylists(ctx context.Context) ([]model.Playlist, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT
	p.id,
	p.name,
	p.m3u_url,
	p.epg_url,
	p.update_interval_minutes,
	p.enabled,
	(SELECT COUNT(*) FROM channels c WHERE c.playlist_id = p.id) AS channel_count,
	(SELECT COUNT(DISTINCT pr.channel_id) FROM programs pr WHERE pr.playlist_id = p.id) AS epg_channel_count,
	p.last_sync_at,
	p.last_error
FROM playlists p
ORDER BY p.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Playlist
	for rows.Next() {
		var p model.Playlist
		var enabled int
		var syncAt sql.NullString
		var lastErr sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.M3UURL, &p.EPGURL, &p.UpdateIntervalMinutes, &enabled, &p.ChannelCount, &p.EPGChannelCount, &syncAt, &lastErr); err != nil {
			return nil, err
		}
		p.Enabled = enabled == 1
		if syncAt.Valid {
			t, _ := time.Parse(time.RFC3339, syncAt.String)
			p.LastSyncAt = &t
		}
		p.LastError = lastErr.String
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) UpdatePlaylist(ctx context.Context, id int64, p model.Playlist) error {
	_, err := s.db.ExecContext(ctx, `UPDATE playlists SET name=?,m3u_url=?,epg_url=?,update_interval_minutes=?,enabled=? WHERE id=?`, p.Name, p.M3UURL, p.EPGURL, p.UpdateIntervalMinutes, boolToInt(p.Enabled), id)
	return err
}

func (s *Store) DeletePlaylist(ctx context.Context, id int64) error {
	_, _ = s.db.ExecContext(ctx, `DELETE FROM programs WHERE playlist_id=?`, id)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM hidden_groups WHERE playlist_id=?`, id)
	_, _ = s.db.ExecContext(ctx, `DELETE FROM hidden_channels WHERE playlist_id=?`, id)
	_, err := s.db.ExecContext(ctx, `DELETE FROM playlists WHERE id=?`, id)
	if err != nil {
		return err
	}
	_, _ = s.db.ExecContext(ctx, `DELETE FROM channels WHERE playlist_id=?`, id)
	return nil
}

func (s *Store) GetPlaylist(ctx context.Context, id int64) (model.Playlist, error) {
	var p model.Playlist
	var enabled int
	var syncAt sql.NullString
	var lastErr sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT
	p.id,
	p.name,
	p.m3u_url,
	p.epg_url,
	p.update_interval_minutes,
	p.enabled,
	(SELECT COUNT(*) FROM channels c WHERE c.playlist_id = p.id) AS channel_count,
	(SELECT COUNT(DISTINCT pr.channel_id) FROM programs pr WHERE pr.playlist_id = p.id) AS epg_channel_count,
	p.last_sync_at,
	p.last_error
FROM playlists p
WHERE p.id=?`, id).
		Scan(&p.ID, &p.Name, &p.M3UURL, &p.EPGURL, &p.UpdateIntervalMinutes, &enabled, &p.ChannelCount, &p.EPGChannelCount, &syncAt, &lastErr)
	if err != nil {
		return model.Playlist{}, err
	}
	p.Enabled = enabled == 1
	if syncAt.Valid {
		t, _ := time.Parse(time.RFC3339, syncAt.String)
		p.LastSyncAt = &t
	}
	p.LastError = lastErr.String
	return p, nil
}

func (s *Store) ReplaceChannels(ctx context.Context, playlistID int64, channels []model.Channel) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM programs WHERE playlist_id=?`, playlistID); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(channels))
	for _, c := range channels {
		seen[c.ExternalID] = struct{}{}
		_, err := tx.ExecContext(ctx, `
INSERT INTO channels(playlist_id,external_id,name,group_name,sort_index,logo,stream_url,archive_supported,stream_mode_cache,stream_mode_checked_at)
VALUES(?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(playlist_id, external_id) DO UPDATE SET
	name=excluded.name,
	group_name=excluded.group_name,
	sort_index=excluded.sort_index,
	logo=excluded.logo,
	stream_url=excluded.stream_url,
	archive_supported=excluded.archive_supported,
	stream_mode_cache=NULL,
	stream_mode_checked_at=NULL`,
			playlistID, c.ExternalID, c.Name, c.Group, c.SortIndex, c.Logo, c.StreamURL, boolToInt(c.ArchiveSupported), nil, nil)
		if err != nil {
			return err
		}
	}
	if len(seen) > 0 {
		holders := make([]string, 0, len(seen))
		args := make([]any, 0, len(seen)+1)
		args = append(args, playlistID)
		for extID := range seen {
			holders = append(holders, "?")
			args = append(args, extID)
		}
		q := fmt.Sprintf(`DELETE FROM channels WHERE playlist_id=? AND external_id NOT IN (%s)`, strings.Join(holders, ","))
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListChannelsByPlaylist(ctx context.Context, playlistID int64) ([]model.Channel, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,playlist_id,external_id,name,group_name,sort_index,logo,stream_url,archive_supported,stream_mode_cache,stream_mode_checked_at FROM channels WHERE playlist_id=? ORDER BY CASE WHEN sort_index>0 THEN 0 ELSE 1 END, sort_index, id`, playlistID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Channel
	for rows.Next() {
		var c model.Channel
		var archived int
		var mode sql.NullString
		var modeAt sql.NullString
		if err := rows.Scan(&c.ID, &c.PlaylistID, &c.ExternalID, &c.Name, &c.Group, &c.SortIndex, &c.Logo, &c.StreamURL, &archived, &mode, &modeAt); err != nil {
			return nil, err
		}
		c.ArchiveSupported = archived == 1
		c.StreamModeCache = mode.String
		if modeAt.Valid {
			t, _ := time.Parse(time.RFC3339, modeAt.String)
			c.StreamModeAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) ListChannelsByPlaylistPage(ctx context.Context, playlistID int64, limit int, offset int) ([]model.Channel, error) {
	if limit <= 0 {
		return s.ListChannelsByPlaylist(ctx, playlistID)
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,playlist_id,external_id,name,group_name,sort_index,logo,stream_url,archive_supported,stream_mode_cache,stream_mode_checked_at
FROM channels
WHERE playlist_id=?
ORDER BY CASE WHEN sort_index>0 THEN 0 ELSE 1 END, sort_index, id
LIMIT ? OFFSET ?`, playlistID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Channel
	for rows.Next() {
		var c model.Channel
		var archived int
		var mode sql.NullString
		var modeAt sql.NullString
		if err := rows.Scan(&c.ID, &c.PlaylistID, &c.ExternalID, &c.Name, &c.Group, &c.SortIndex, &c.Logo, &c.StreamURL, &archived, &mode, &modeAt); err != nil {
			return nil, err
		}
		c.ArchiveSupported = archived == 1
		c.StreamModeCache = mode.String
		if modeAt.Valid {
			t, _ := time.Parse(time.RFC3339, modeAt.String)
			c.StreamModeAt = &t
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) GetChannel(ctx context.Context, id int64) (model.Channel, error) {
	var c model.Channel
	var archived int
	var mode sql.NullString
	var modeAt sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT id,playlist_id,external_id,name,group_name,sort_index,logo,stream_url,archive_supported,stream_mode_cache,stream_mode_checked_at FROM channels WHERE id=?`, id).Scan(&c.ID, &c.PlaylistID, &c.ExternalID, &c.Name, &c.Group, &c.SortIndex, &c.Logo, &c.StreamURL, &archived, &mode, &modeAt)
	if err != nil {
		return model.Channel{}, err
	}
	c.ArchiveSupported = archived == 1
	c.StreamModeCache = mode.String
	if modeAt.Valid {
		t, _ := time.Parse(time.RFC3339, modeAt.String)
		c.StreamModeAt = &t
	}
	return c, nil
}

func (s *Store) SetChannelStreamModeCache(ctx context.Context, channelID int64, mode string, checkedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE channels SET stream_mode_cache=?, stream_mode_checked_at=? WHERE id=?`, mode, checkedAt.UTC().Format(time.RFC3339), channelID)
	return err
}

func (s *Store) GetHiddenPrefsByPlaylist(ctx context.Context, playlistID int64) ([]string, []string, error) {
	groupRows, err := s.db.QueryContext(ctx, `SELECT group_name FROM hidden_groups WHERE playlist_id=? ORDER BY group_name`, playlistID)
	if err != nil {
		return nil, nil, err
	}
	defer groupRows.Close()
	groups := make([]string, 0, 32)
	for groupRows.Next() {
		var name string
		if err := groupRows.Scan(&name); err != nil {
			return nil, nil, err
		}
		groups = append(groups, name)
	}
	if err := groupRows.Err(); err != nil {
		return nil, nil, err
	}

	channelRows, err := s.db.QueryContext(ctx, `SELECT channel_external_id FROM hidden_channels WHERE playlist_id=? ORDER BY channel_external_id`, playlistID)
	if err != nil {
		return nil, nil, err
	}
	defer channelRows.Close()
	channels := make([]string, 0, 64)
	for channelRows.Next() {
		var externalID string
		if err := channelRows.Scan(&externalID); err != nil {
			return nil, nil, err
		}
		channels = append(channels, externalID)
	}
	if err := channelRows.Err(); err != nil {
		return nil, nil, err
	}
	return groups, channels, nil
}

func (s *Store) ReplaceHiddenPrefsByPlaylist(ctx context.Context, playlistID int64, groups []string, channelExternalIDs []string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM hidden_groups WHERE playlist_id=?`, playlistID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM hidden_channels WHERE playlist_id=?`, playlistID); err != nil {
		return err
	}
	for _, group := range groups {
		name := strings.TrimSpace(group)
		if name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO hidden_groups(playlist_id, group_name) VALUES(?, ?)`, playlistID, name); err != nil {
			return err
		}
	}
	for _, externalID := range channelExternalIDs {
		id := strings.TrimSpace(externalID)
		if id == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO hidden_channels(playlist_id, channel_external_id) VALUES(?, ?)`, playlistID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) SetChannelLogo(ctx context.Context, channelID int64, logo string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE channels SET logo=? WHERE id=?`, logo, channelID)
	return err
}

func (s *Store) ReplacePrograms(ctx context.Context, channelID int64, programs []model.Program) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM programs WHERE channel_id=?`, channelID); err != nil {
		return err
	}
	for _, p := range programs {
		_, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO programs(playlist_id,channel_id,title,start_at,end_at,description) VALUES(?,?,?,?,?,?)`, p.PlaylistID, channelID, p.Title, p.StartAt.UTC().Format(time.RFC3339), p.EndAt.UTC().Format(time.RFC3339), p.Description)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) GetProgramsByWindow(ctx context.Context, channelID int64, from, to time.Time) ([]model.Program, error) {
	var playlistID int64
	err := s.db.QueryRowContext(ctx, `SELECT playlist_id FROM channels WHERE id=?`, channelID).Scan(&playlistID)
	if err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT playlist_id,channel_id,title,start_at,end_at,description FROM programs WHERE playlist_id=? AND channel_id=? AND end_at>=? AND start_at<=? ORDER BY start_at`, playlistID, channelID, from.UTC().Format(time.RFC3339), to.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Program
	for rows.Next() {
		var p model.Program
		var start, end string
		if err := rows.Scan(&p.PlaylistID, &p.ChannelID, &p.Title, &start, &end, &p.Description); err != nil {
			return nil, err
		}
		p.StartAt, _ = time.Parse(time.RFC3339, start)
		p.EndAt, _ = time.Parse(time.RFC3339, end)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetNowProgramsByPlaylist(ctx context.Context, playlistID int64, at time.Time) ([]model.Program, error) {
	ts := at.UTC().Format(time.RFC3339)
	rows, err := s.db.QueryContext(ctx, `
SELECT p.playlist_id, p.channel_id, p.title, p.start_at, p.end_at, p.description
FROM programs p
WHERE p.playlist_id = ?
  AND p.start_at <= ?
  AND p.end_at > ?
  AND p.start_at = (
    SELECT MAX(p2.start_at)
    FROM programs p2
    WHERE p2.playlist_id = p.playlist_id
      AND p2.channel_id = p.channel_id
      AND p2.start_at <= ?
      AND p2.end_at > ?
  )
ORDER BY p.channel_id
`, playlistID, ts, ts, ts, ts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Program
	for rows.Next() {
		var p model.Program
		var start, end string
		if err := rows.Scan(&p.PlaylistID, &p.ChannelID, &p.Title, &start, &end, &p.Description); err != nil {
			return nil, err
		}
		p.StartAt, _ = time.Parse(time.RFC3339, start)
		p.EndAt, _ = time.Parse(time.RFC3339, end)
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) SetPlaylistSyncStatus(ctx context.Context, playlistID int64, lastErr string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.ExecContext(ctx, `UPDATE playlists SET last_sync_at=?, last_error=? WHERE id=?`, now, lastErr, playlistID)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
