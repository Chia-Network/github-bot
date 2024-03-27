package config

// Config defines the config for all aspects of the bot
type Config struct {
	GithubToken     string   `yaml:"github_token"`
	InternalTeam    string   `yaml:"internal_team"`
	LabelCheckRepos []string `yaml:"label_check_repos"`
	LabelSkipUsers  []string `yaml:"label_skip_users"`
	LabelSkipMap    map[string]bool
}
