package config

import (
	"os"
	"strconv"
)

type Config struct {
	Addr                  string
	DBPath                string
	LogLevel              string
	AppTitle              string
	SingleConnPolicy      string
	FFmpegBin             string
	StreamIdleTimeoutSec  int
	StreamRetryMax        int
	StreamModeCacheTTLMin int
	ManifestBackoffEnable bool
}

func Load() Config {
	return Config{
		Addr:                  getEnv("WEBTV_ADDR", ":8080"),
		DBPath:                getEnv("WEBTV_DB_PATH", "webtv.db"),
		LogLevel:              getEnv("WEBTV_LOG_LEVEL", "info"),
		AppTitle:              getEnv("WEBTV_APP_TITLE", "WebTV"),
		SingleConnPolicy:      getEnv("WEBTV_SINGLE_CONN_POLICY", "reject"),
		FFmpegBin:             getEnv("WEBTV_FFMPEG_BIN", "ffmpeg"),
		StreamIdleTimeoutSec:  getEnvInt("WEBTV_STREAM_IDLE_TIMEOUT_SEC", 12),
		StreamRetryMax:        getEnvInt("WEBTV_STREAM_RETRY_MAX", 5),
		StreamModeCacheTTLMin: getEnvInt("WEBTV_STREAM_MODE_CACHE_TTL_MIN", 360),
		ManifestBackoffEnable: getEnvBool("WEBTV_STREAM_MANIFEST_BACKOFF_ENABLED", false),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func getEnvBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}
