package state

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

// Repository keeps track of all active project + ref state
type Repository struct {
	// Config tracks the config that contains definitions for all supported projects and the required/trigger workflows for them
	Configs *config.Configs

	// Projects - List of supported projects for quick lookup
	Projects map[string]bool

	// githubToken used to authenticate to GitHub API
	githubToken string

	mysqlClient *sql.DB
}

// GitHubRequestData structure that data must be passed to github
type GitHubRequestData struct {
	Ref    string            `json:"ref"`
	Inputs map[string]string `json:"inputs"`
}

// NewRepository returns a new repository object
func NewRepository(configs *config.Configs, githubToken string) (*Repository, error) {
	repo := &Repository{
		Configs:     configs,
		Projects:    map[string]bool{},
		githubToken: githubToken,
	}

	// Initialize all supported projects
	for _, cfg := range *configs {
		repo.Projects[cfg.Name] = true
	}

	err := repo.createDBClient()
	if err != nil {
		return nil, fmt.Errorf("error creating mysql client: %w", err)
	}
	err = repo.initTables()
	if err != nil {
		return nil, fmt.Errorf("error ensuring tables exist in MySQL. %w", err)
	}

	return repo, nil
}

func (r *Repository) createDBClient() error {
	var err error

	cfg := mysql.Config{
		User:                 viper.GetString("mysql-user"),
		Passwd:               viper.GetString("mysql-password"),
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", viper.GetString("mysql-host"), viper.GetUint16("mysql-port")),
		DBName:               viper.GetString("mysql-db-name"),
		AllowNativePasswords: true,
	}
	r.mysqlClient, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}

	r.mysqlClient.SetConnMaxLifetime(time.Minute * 3)
	r.mysqlClient.SetMaxOpenConns(10)
	r.mysqlClient.SetMaxIdleConns(10)

	return nil
}
