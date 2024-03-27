package config

// Config defines the config for all aspects of the bot
type Config struct {
	GithubToken  string `yaml:"github_token"`
	InternalTeam string `yaml:"internal_team"`
	LabelConfig
}

// LabelConfig is the configuration options specific to labeling PRs
type LabelConfig struct {
	LabelInternal      string   `yaml:"label_internal"`
	LabelExternal      string   `yaml:"label_external"`
	LabelCheckRepos    []string `yaml:"label_check_repos"`
	LabelSkipUsers     []string `yaml:"label_skip_users"`
	LabelMinimumNumber int      `yaml:"label_minimum_number"`

	LabelSkipMap map[string]bool
}
