<p align="center">
  <img src="web/public/webtv-logo.svg" alt="WebTV logo" width="120" height="120">
</p>

# WebTV

WebTV is a small self-hosted IPTV player that runs in the browser. Add one or more M3U playlists, open the web UI from any computer on your network, and watch TV without installing a separate player on every machine.

This project is just for fun and was written with help from LLMs. I simply wanted a browser-based IPTV video player for my own use: something I could run once on a server or desktop and then open from any browser.

[Русская версия](README_RU.md)

## Features

- Browser-based IPTV playback with a bundled Vue web UI.
- Multiple M3U playlists with names, enable/disable state, manual refresh, and automatic refresh intervals.
- EPG/XMLTV import with current and upcoming programs.
- Channel groups, channel search, favorites, and per-playlist hidden groups/channels.
- Archive/timeshift support when a playlist exposes archive/catchup metadata.
- Automatic stream mode selection: direct proxy when the browser can play the stream, ffmpeg transcoding when it cannot.
- HLS playback through hls.js where needed, with native playback fallback.
- Audio track discovery and selection, including a default audio language preference.
- Optional deinterlacing for interlaced IPTV streams.
- Light, dark, and system UI themes.
- English and Russian localization.
- SQLite storage in a single database file.

## How Playback and Transcoding Work

WebTV does not blindly transcode every channel. It tries to keep streams as close to the original as possible:

- **Direct mode** proxies playable streams to the browser. For HLS playlists, WebTV rewrites `.m3u8` segment URLs so the browser can load nested segments through the same server.
- **Transcode mode** uses `ffmpeg` and outputs fragmented MP4 (`video/mp4`) for broad HTML5 video compatibility.
- The server probes HLS/codec information and switches to transcoding for browser-unfriendly codecs such as HEVC/H.265, MPEG-2 video, MPEG-4 video, AC3/EAC3/A52, MP2, and MPGA.
- When possible, ffmpeg copies compatible video or audio streams and only re-encodes the incompatible part.
- If deinterlacing is enabled in the UI, video is decoded and encoded with the `yadif` filter.
- If a selected audio track does not work through the browser/HLS path, the UI can request an ffmpeg audio fallback.
- Stream mode detection is cached per channel for `WEBTV_STREAM_MODE_CACHE_TTL_MIN` minutes.
- Direct stream proxying has an idle timeout and retry loop for unstable IPTV sources.
- HLS manifests are cached briefly; optional manifest backoff can reduce reload pressure on providers that return unchanged live playlists.

The Docker image already includes `ffmpeg`, `ffprobe`, CA certificates, and timezone data. If you run the binary directly, install `ffmpeg` separately and make sure `ffmpeg` and `ffprobe` are available in `PATH`, or configure `WEBTV_FFMPEG_BIN`.

## Configuration

Configuration is done with environment variables.

| Variable | Default | Description |
| --- | --- | --- |
| `WEBTV_ADDR` | `:8080` | HTTP listen address. Use values like `:8080` or `127.0.0.1:8080`. |
| `WEBTV_DB_PATH` | `webtv.db` | SQLite database file path. In Docker Compose the default is `/data/webtv.db`, with `/data` mounted to the local `./data` directory. |
| `WEBTV_LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, or `error`. |
| `WEBTV_APP_TITLE` | `WebTV` | Title shown in the web UI header. |
| `WEBTV_SINGLE_CONN_POLICY` | `reject` | Historical stream connection policy option. Current stream handling allows browser retry/reopen behavior, so this is mostly kept for compatibility. |
| `WEBTV_FFMPEG_BIN` | `ffmpeg` | ffmpeg executable name or absolute path used for transcoding. `ffprobe` is also used for probing. |
| `WEBTV_STREAM_IDLE_TIMEOUT_SEC` | `12` | Idle timeout for direct stream copying. If no data arrives during this period, WebTV reconnects. |
| `WEBTV_STREAM_RETRY_MAX` | `5` | Maximum reconnect attempts for stalled direct streams. |
| `WEBTV_STREAM_MODE_CACHE_TTL_MIN` | `360` | How long detected direct/transcode mode is cached for a channel. |
| `WEBTV_STREAM_MANIFEST_BACKOFF_ENABLED` | `false` | Enables short HLS manifest backoff when live manifests are unchanged. |
| `WEBTV_IMAGE_TAG` | `latest` | Docker Compose image tag for `ghcr.io/asd2003ru/webtv`. |
| `WEBTV_PORT` | `8080` | Host port mapped to container port `8080` in `docker-compose.yaml`. |

See `.env-example` for a ready-to-edit example.

## Install with Docker

Docker is the main and recommended way to install WebTV. A ready-to-use `docker-compose.yaml` is included in the repository, so you only need to download it and prepare an `.env` file.

Download these files from the repository:

- `docker-compose.yaml`
- `.env-example`

Then run:

```bash
cp .env-example .env
docker compose up -d
```

Then open:

```text
http://localhost:8080
```

Data is stored in the local `./data` directory next to `docker-compose.yaml`. WebTV creates `data/webtv.db` automatically on first start.

Minimal one-command run:

```bash
docker run -d \
  --name webtv \
  --restart unless-stopped \
  -p 8080:8080 \
  -v ./data:/data \
  ghcr.io/asd2003ru/webtv:latest
```

## Install from Binary

Download or build the `webtv` binary, then run it:

```bash
WEBTV_ADDR=:8080 \
WEBTV_DB_PATH=./data/webtv.db \
WEBTV_FFMPEG_BIN=ffmpeg \
./webtv
```

Requirements for binary mode:

- `ffmpeg` and `ffprobe` installed.
- A modern browser on client devices.
- Network access from the WebTV host to your IPTV provider URLs.

## Build

Build the binary:

```bash
make install
make build
```

The binary will be created at:

```text
bin/webtv
```

Equivalent manual steps:

```bash
npm --prefix web install
npm --prefix web run build
go build -o bin/webtv ./cmd/webtv
```

Run from source:

```bash
make run
```

Run tests:

```bash
make test
```

## Build the Docker Image

Build with Docker Compose:

```bash
docker compose -f docker-compose-build.yaml up -d --build
```

Or build the image directly:

```bash
docker build -t webtv:latest .
```

The Dockerfile builds the Vue frontend first, embeds it into the Go binary, and then copies the binary into a small Alpine runtime image with ffmpeg installed.

## Adding a New UI Language

Localization lives in `web/src/i18n/messages.js`, and language selection logic lives in `web/src/composables/useUILanguage.js`.

To add a language:

1. Add a new top-level language key to `messages`, for example `de`.
2. Copy all keys from `en` and translate every value. Keep the same key names.
3. Update `systemLanguageCode()` in `useUILanguage.js` if the new language should be auto-detected from `navigator.language`.
4. Allow the new language mode in `currentLanguage()`, `onLanguageModeChange()`, and `initLanguageMode()`.
5. Add a visible option in `web/src/components/PlaylistsPanel.vue` for the new language.
6. Add a label key such as `lang_german` to every language block.
7. Run `npm --prefix web run build` and check the UI.

## Notes

- The app is intended for personal/self-hosted use.
- WebTV does not provide IPTV content. You need your own legal M3U/EPG sources.
- Transcoding is CPU-intensive. If many clients watch transcoded channels at once, run WebTV on hardware with enough CPU headroom.
