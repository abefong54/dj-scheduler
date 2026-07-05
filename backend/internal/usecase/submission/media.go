package submission

import (
	"net/url"
	"strings"

	"eventlineup/internal/domain/apperrors"
)

// Media is the resolved reference to an embedded player: the provider and the
// minimal id/URL we need to build a known-good iframe on the frontend. We never
// store or trust provider HTML — only these two extracted values.
type Media struct {
	Provider string // youtube | soundcloud | mixcloud
	Ref      string // youtube: video id; soundcloud/mixcloud: normalized track URL
}

// ResolveMedia is the host-allowlist security boundary for a pasted link. It
// parses (never fetches) the URL, confirms the host is one of the three allowed
// providers, and extracts (provider, ref). Anything else is rejected as a
// validation error so it surfaces to the student as a 400, not a 500.
func ResolveMedia(raw string) (Media, error) {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return Media{}, apperrors.NewValidation("invalid media URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return Media{}, apperrors.NewValidation("media URL must be http(s)")
	}
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")

	switch host {
	case "youtube.com", "m.youtube.com", "music.youtube.com":
		if v := u.Query().Get("v"); v != "" {
			return Media{"youtube", v}, nil
		}
		// /embed/<id> or /shorts/<id> or /live/<id>
		if strings.HasPrefix(u.Path, "/embed/") || strings.HasPrefix(u.Path, "/shorts/") || strings.HasPrefix(u.Path, "/live/") {
			if id := lastPathSegment(u.Path); id != "" {
				return Media{"youtube", id}, nil
			}
		}
		return Media{}, apperrors.NewValidation("could not find a YouTube video id in that link")
	case "youtu.be":
		if id := lastPathSegment(u.Path); id != "" {
			return Media{"youtube", id}, nil
		}
		return Media{}, apperrors.NewValidation("could not find a YouTube video id in that link")
	case "soundcloud.com":
		if strings.Trim(u.Path, "/") == "" {
			return Media{}, apperrors.NewValidation("that SoundCloud link has no track")
		}
		return Media{"soundcloud", "https://soundcloud.com" + u.Path}, nil
	case "mixcloud.com":
		if strings.Trim(u.Path, "/") == "" {
			return Media{}, apperrors.NewValidation("that Mixcloud link has no show")
		}
		return Media{"mixcloud", "https://www.mixcloud.com" + u.Path}, nil
	default:
		return Media{}, apperrors.NewValidation("unsupported provider — paste a YouTube, SoundCloud, or Mixcloud link")
	}
}

// lastPathSegment returns the final non-empty segment of a URL path, or "".
func lastPathSegment(p string) string {
	segs := strings.Split(strings.Trim(p, "/"), "/")
	if len(segs) == 0 {
		return ""
	}
	return segs[len(segs)-1]
}
