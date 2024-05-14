package config

// Config defines the config for all aspects of the bot
type Config struct {
	GithubToken       string `yaml:"github_token"`
	InternalTeam      string `yaml:"internal_team"`
	LabelConfig       `yaml:",inline"`
	CheckStalePending `yaml:",inline"`
}

// LabelConfig is the configuration options specific to labeling PRs
type LabelConfig struct {
	LabelInternal   string      `yaml:"label_internal"`
	LabelExternal   string      `yaml:"label_external"`
	LabelCheckRepos []CheckRepo `yaml:"label_check_repos"`
	LabelSkipUsers  []string    `yaml:"label_skip_users"`
	LabelSkipMap    map[string]bool
}

// CheckRepo is config settings when checking a repo
type CheckRepo struct {
	Name          string `yaml:"name"`
	MinimumNumber int    `yaml:"minimum_number"`
}

// CheckStalePending are config settings when checking a repo
type CheckStalePending struct {
	CheckStalePending []CheckRepo `yaml:"check_stale_pending_repos"`
}
