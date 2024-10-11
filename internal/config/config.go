package config

// Config defines the config for all aspects of the bot
type Config struct {
	GithubToken              string   `yaml:"github_token"`
	InternalTeam             string   `yaml:"internal_team"`
	InternalTeamIgnoredUsers []string `yaml:"internal_team_ignored_users"`
	SkipUsers                []string `yaml:"skip_users"`
	SkipUsersMap             map[string]bool
	LabelConfig              `yaml:",inline"`
	CheckRepos               []CheckRepo `yaml:"check_repos"`
}

// LabelConfig is the configuration options specific to labeling PRs
type LabelConfig struct {
	LabelInternal string `yaml:"label_internal"`
	LabelExternal string `yaml:"label_external"`
}

// CheckRepo is config settings when checking a repo
type CheckRepo struct {
	Name          string `yaml:"name"`
	MinimumNumber int    `yaml:"minimum_number"`
}
