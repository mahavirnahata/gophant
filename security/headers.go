package security

import "net/http"

type HeaderConfig struct {
	ContentTypeNoSniff bool
	FrameDeny          bool
	ReferrerPolicy     string
	HSTS               bool
	HSTSMaxAge         int
}

func DefaultHeaders() HeaderConfig {
	return HeaderConfig{
		ContentTypeNoSniff: true,
		FrameDeny:          true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
		HSTS:               false,
		HSTSMaxAge:         63072000,
	}
}

func ApplyHeaders(w http.ResponseWriter, cfg HeaderConfig) {
	if cfg.ContentTypeNoSniff {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}
	if cfg.FrameDeny {
		w.Header().Set("X-Frame-Options", "DENY")
	}
	if cfg.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
	}
	if cfg.HSTS {
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	}
}
