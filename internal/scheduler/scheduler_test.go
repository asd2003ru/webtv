package scheduler

import (
	"testing"
	"webtv/internal/model"
)

func TestNormalizeRelaxedChannelName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "Первый канал", want: "первый"},
		{in: "Первый канал HD", want: "первый"},
		{in: "Discovery Channel UHD", want: "discovery channel"},
		{in: "  Россия 1  ", want: "россия 1"},
	}
	for _, tt := range tests {
		got := normalizeRelaxedChannelName(tt.in)
		if got != tt.want {
			t.Fatalf("normalizeRelaxedChannelName(%q)=%q want %q", tt.in, got, tt.want)
		}
	}
}

func TestPickBestChannelCandidate_PrefersHDWhenEPGHasHD(t *testing.T) {
	channels := map[int64]model.Channel{
		1: {ID: 1, Name: "Первый канал"},
		2: {ID: 2, Name: "Первый канал HD"},
	}
	got := pickBestChannelCandidate([]int64{1, 2}, channels, "Первый канал", []string{"Первый канал HD"})
	if got != 2 {
		t.Fatalf("expected HD channel id 2, got %d", got)
	}
}

func TestBaseNameKey_StripsTechnicalSuffixButKeepsTimeshift(t *testing.T) {
	if got := baseNameKey("Первый канал HD"); got != "первый канал" {
		t.Fatalf("expected base name without hd, got %q", got)
	}
	if got := baseNameKey("Первый канал +2"); got != "первый канал 2" {
		t.Fatalf("expected timeshift name kept, got %q", got)
	}
}

func TestFindEPGDonorChannelID(t *testing.T) {
	byName := map[string][]int64{
		"первый канал":         {1},
		"первый канал hd":      {2},
		"первый канал hd orig": {3},
	}
	chByID := map[int64]model.Channel{
		1: {ID: 1, Name: "Первый канал"},
		2: {ID: 2, Name: "Первый канал HD"},
		3: {ID: 3, Name: "Первый канал HD orig"},
	}
	chPrograms := map[int64][]model.Program{
		1: {{Title: "Новости"}},
	}
	if got := findEPGDonorChannelID("Первый канал HD orig", byName, chByID, chPrograms); got != 1 {
		t.Fatalf("expected donor id 1, got %d", got)
	}
	if got := findEPGDonorChannelID("Первый канал +2", byName, chByID, chPrograms); got != 0 {
		t.Fatalf("expected no donor for +2 channel, got %d", got)
	}
}

func TestDonorNameKey(t *testing.T) {
	if got := donorNameKey("Россия 1 HD"); got != "россия 1" {
		t.Fatalf("unexpected donor key: %q", got)
	}
	if got := donorNameKey("Россия 1 orig"); got != "россия 1" {
		t.Fatalf("unexpected donor key: %q", got)
	}
	if got := donorNameKey("ТНТ +2"); got != "" {
		t.Fatalf("timeshift channel should not have donor key, got %q", got)
	}
}
