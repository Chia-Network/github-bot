package config

// Config defines the config for all aspects of the bot
type Config struct {
	GithubToken       string `yaml:"github_token"`
	InternalTeam      string `yaml:"internal_team"`
	LabelConfig       `yaml:",inline"`
	CheckStalePending `yaml:",inline"`
	CheckRepo         `yaml:",inline"`
}

// LabelConfig is the configuration options specific to labeling PRs
type LabelConfig struct {
	LabelInternal   string      `yaml:"label_internal"`
	LabelExternal   string      `yaml:"label_external"`
	LabelCheckRepos []CheckRepo `yaml:"label_check_repos"`
}

// CheckRepo is config settings when checking a repo
type CheckRepo struct {
	Name          string   `yaml:"name"`
	MinimumNumber int      `yaml:"minimum_number"`
	SkipUsers     []string `yaml:"label_skip_users"`
	SkipUsersMap  map[string]bool
}

// CheckStalePending are config settings when checking a repo
type CheckStalePending struct {
	CheckStalePending []CheckRepo `yaml:"check_stale_pending_repos"`
}
