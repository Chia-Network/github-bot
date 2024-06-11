package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

// PRInfo struct holds the PR information
type PRInfo struct {
	Project         string
	PRNumber        int64
	LastMessageSent time.Time
}

// Datastore manages connections and the state of the database.
type Datastore struct {
	mysqlClient *sql.DB
	dbHost      string
	dbPort      uint16
	dbUser      string
	dbPass      string
	dbName      string
}

// NewDatastore initializes a new Datastore with the given configurations.
func NewDatastore(dbHost string, dbPort uint16, dbUser string, dbPass string, dbName string) (*Datastore, error) {

	datastore := &Datastore{
		dbHost: dbHost,
		dbPort: dbPort,
		dbUser: dbUser,
		dbPass: dbPass,
		dbName: dbName,
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

func (d *Datastore) GetPRInfo(project string, prNumber int64) (*PRInfo, error) {
	// Prepare the query to fetch the PR information
	query := "SELECT project, pr_number, last_message_sent FROM project_status WHERE project = ? AND pr_number = ?"

	// Variable to store the results
	var prInfo PRInfo

	// Execute the query
	err := d.mysqlClient.QueryRow(query, project, prNumber).Scan(&prInfo.Project, &prInfo.PRNumber, &prInfo.LastMessageSent)
	if err != nil {
		if err == sql.ErrNoRows {
			// Handle no rows returned case here if needed
			return nil, fmt.Errorf("no data found for project %s and PR number %d", project, prNumber)
		}
		// Handle other errors
		return nil, fmt.Errorf("error querying PR info: %v", err)
	}

	// Return the fetched data
	return &prInfo, nil
}

func (d *Datastore) StoreAuditData(project string, prNumber int64) error {
	result, err := d.mysqlClient.Query("INSERT INTO project_status (project, pr_number, last_message_sent) VALUES (?, ?, NOW())"+
		" ON DUPLICATE KEY UPDATE last_message_sent = VALUES(last_message_sent);", project, prNumber)
	if err != nil {
		return fmt.Errorf("error inserting or updating project status: %v", err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Could not close rows: %s\n", err.Error())
		}
	}(result)

	return err
}
