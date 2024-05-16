package keybase

import (
	"fmt"
)

func (d *Datastore) initTables() error {
	if d.mysqlClient == nil {
		return fmt.Errorf("mysqlClient not initialized")
	}
	query := "CREATE TABLE IF NOT EXISTS `projects` (" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"  `project` VARCHAR(255) NOT NULL," +
		"  `prnumber` bigint NOT NULL," +
		"  `msg_last_sent_time` DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP," +
		"  PRIMARY KEY (`id`)," +
		"  UNIQUE KEY `project-pr-number-unique` (`project`, `prnumber`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"

	result, err := d.mysqlClient.Query(query)
	if err != nil {
		return err
	}
	err = result.Close()
	if err != nil {
		return err
	}

	return nil
}
