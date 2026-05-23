package gutter

import "github.com/Runway-Club/gutter/themes"

// Option configures a RunApp invocation. Use the With* functions to construct
// options rather than building the struct directly.
type Option func(*runConfig)

// runConfig is the private aggregation of all options for one RunApp call.
type runConfig struct {
	selector string
	theme    *themes.Theme
}

func newRunConfig(opts []Option) *runConfig {
	cfg := &runConfig{
		selector: "#app",
		theme:    themes.Apple,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// WithTheme sets the theme that will populate BuildContext.Theme. Themed
// widgets read from it; raw widgets ignore it.
func WithTheme(t *themes.Theme) Option {
	return func(c *runConfig) {
		if t != nil {
			c.theme = t
		}
	}
}

// WithSelector overrides the CSS selector used to find the mount point.
// Defaults to "#app".
func WithSelector(s string) Option {
	return func(c *runConfig) {
		if s != "" {
			c.selector = s
		}
	}
}
