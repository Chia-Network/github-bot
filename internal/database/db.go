package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/chia-network/go-modules/pkg/slogs"
	"github.com/go-sql-driver/mysql"
)

// PRInfo struct holds the PR information
type PRInfo struct {
	Repo             string
	PRNumber         int64
	LastMessageSent  time.Time
	SuppressMessages bool
}

// Datastore manages connections and the state of the database.
type Datastore struct {
	mysqlClient *sql.DB
	dbHost      string
	dbPort      uint16
	dbUser      string
	dbPass      string
	dbName      string
	tableName   string
}

// NewDatastore initializes a new Datastore with the given configurations.
func NewDatastore(dbHost string, dbPort uint16, dbUser string, dbPass string, dbName string, tableName string) (*Datastore, error) {

	datastore := &Datastore{
		dbHost:    dbHost,
		dbPort:    dbPort,
		dbUser:    dbUser,
		dbPass:    dbPass,
		dbName:    dbName,
		tableName: tableName,
	}

	err := datastore.createDBClient()
	if err != nil {
		return nil, fmt.Errorf("error creating mysql client: %w", err)
	}
	err = datastore.initTables()
	if err != nil {
		return nil, fmt.Errorf("error ensuring tables exist in MySQL: %w", err)
	}

	return datastore, nil
}

// createDBClient sets up the database connection.
func (d *Datastore) createDBClient() error {
	var err error
	cfg := mysql.Config{
		User:                 d.dbUser,
		Passwd:               d.dbPass,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", d.dbHost, d.dbPort),
		DBName:               d.dbName,
		AllowNativePasswords: true,
	}
	d.mysqlClient, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return err
	}

	d.mysqlClient.SetConnMaxLifetime(time.Second * 15)
	d.mysqlClient.SetMaxOpenConns(1)
	d.mysqlClient.SetMaxIdleConns(1)

	return nil
}

func (d *Datastore) initTables() error {
	if d.mysqlClient == nil {
		return fmt.Errorf("mysqlClient not initialized")
	}
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` ("+
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT,"+
		"  `repo` VARCHAR(255) NOT NULL,"+
		"  `pr_number` bigint NOT NULL,"+
		"  `last_message_sent` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,"+
		"  `suppress_messages` BOOLEAN NOT NULL DEFAULT FALSE,"+
		"  PRIMARY KEY (`id`),"+
		"  UNIQUE KEY `repo_pr_number_unique` (`repo`, `pr_number`)"+
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;", d.tableName)

	_, err := d.mysqlClient.Exec(query)
	if err != nil {
		return err
	}

	// List of required columns
	requiredColumns := map[string]string{
		"suppress_messages": "BOOLEAN NOT NULL DEFAULT FALSE",
		// Add other columns here as needed
	}

	// Check for missing columns and add them if necessary
	for column, definition := range requiredColumns {
		var columnName string
		query := fmt.Sprintf("SHOW COLUMNS FROM `%s` LIKE '%s'", d.tableName, column)
		err := d.mysqlClient.QueryRow(query).Scan(&columnName, new(string), new(string), new(string), new(string), new(string))
		if err != nil {
			if err == sql.ErrNoRows {
				// Column does not exist, add it
				alterQuery := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN `%s` %s", d.tableName, column, definition)
				_, err := d.mysqlClient.Exec(alterQuery)
				if err != nil {
					return fmt.Errorf("error adding column %s: %v", column, err)
				}
				slogs.Logr.Info("Added column to table", "table", d.tableName, "column", column)
			} else {
				return fmt.Errorf("error checking column %s: %v", column, err)
			}
		}
	}

	return nil
}

// GetPRData retrieves PR information from the database.
func (d *Datastore) GetPRData(repo string, prNumber int64) (*PRInfo, error) {
	// Prepare the query to fetch the PR information
	query := fmt.Sprintf("SELECT repo, pr_number, last_message_sent FROM %s WHERE repo = ? AND pr_number = ?", d.tableName)

	// Variable to store the results
	var prInfo PRInfo
	var lastMessageSentStr string

	// Execute the query
	err := d.mysqlClient.QueryRow(query, repo, prNumber).Scan(&prInfo.Repo, &prInfo.PRNumber, &lastMessageSentStr)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle no rows returned case here if needed
			return nil, nil // No data found is not an error in this context
		}
		// Handle other errors
		return nil, fmt.Errorf("error querying PR info: %v", err)
	}

	// Parse the last_message_sent string to time.Time. Reference date is used here.
	lastMessageSent, err := time.Parse("2006-01-02 15:04:05", lastMessageSentStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing last_message_sent: %v", err)
	}
	prInfo.LastMessageSent = lastMessageSent

	// Return the fetched data
	return &prInfo, nil
}

// StorePRData stores or updates PR information in the database.
func (d *Datastore) StorePRData(repo string, prNumber int64) error {
	query := fmt.Sprintf("INSERT INTO %s (repo, pr_number, last_message_sent) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE last_message_sent = VALUES(last_message_sent);", d.tableName)
	_, err := d.mysqlClient.Exec(query, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error inserting or updating PR status: %v", err)
	}

	return nil
}

// UpdateSuppressMessages updates the suppress_messages flag for a PR.
func (d *Datastore) UpdateSuppressMessages(repo string, prNumber int64, suppress bool) error {
	query := fmt.Sprintf("UPDATE %s SET suppress_messages = ? WHERE repo = ? AND pr_number = ?", d.tableName)
	_, err := d.mysqlClient.Exec(query, suppress, repo, prNumber)
	if err != nil {
		return fmt.Errorf("error updating suppress_messages: %v", err)
	}

	if suppress {
		slogs.Logr.Info("Messages suppressed for PR", "repository", repo, "PR", prNumber)
	} else {
		slogs.Logr.Info("Messages unsuppressed for PR", "repository", repo, "PR", prNumber)
	}

	return nil
}
