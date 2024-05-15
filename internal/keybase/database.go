package keybase

import (
	"fmt"
)

func (r *Repository) initTables() error {
	if r.mysqlClient == nil {
		return fmt.Errorf("mysqlClient not initialized")
	}
	query := "CREATE TABLE IF NOT EXISTS `projects` (" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"  `project` VARCHAR(255) NOT NULL," +
		"  `ref` VARCHAR(255) NOT NULL," +
		"  `vars` TEXT NOT NULL," +
		"  PRIMARY KEY (`id`)," +
		"  UNIQUE KEY `project-ref-unique` (`project`, `ref`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"

	result, err := r.mysqlClient.Query(query)
	if err != nil {
		return err
	}
	err = result.Close()
	if err != nil {
		return err
	}
	query = "CREATE TABLE IF NOT EXISTS `workflow_state` (" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"  `project_id` bigint unsigned NOT NULL ," +
		"  `workflow_name` VARCHAR(255) NOT NULL," +
		"  `complete` TINYINT NOT NULL DEFAULT 0," +
		"  PRIMARY KEY (`id`)," +
		"  KEY `workflow_state_project_id_foreign` (`project_id`)," +
		"  UNIQUE KEY `project-workflow-unique` (`project_id`, `workflow_name`)," +
		"  CONSTRAINT `workflows_state_project_id_foreign` FOREIGN KEY (`project_id`) REFERENCES `projects` (`id`) ON DELETE CASCADE ON UPDATE CASCADE" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"

	result, err = r.mysqlClient.Query(query)
	if err != nil {
		return err
	}
	err = result.Close()
	if err != nil {
		return err
	}

	query = "CREATE TABLE IF NOT EXISTS `delayed_triggers` (" +
		"  `id` bigint unsigned NOT NULL AUTO_INCREMENT," +
		"  `project` VARCHAR(255) NOT NULL," +
		"  `run_after` bigint unsigned NOT NULL," +
		"  `vars` TEXT NOT NULL," +
		"  PRIMARY KEY (`id`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;"

	result, err = r.mysqlClient.Query(query)
	if err != nil {
		return err
	}
	err = result.Close()
	if err != nil {
		return err
	}

	return nil
}
}
