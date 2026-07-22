// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package migration

import (
	"strings"

	"code.vikunja.io/api/pkg/db"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260721120000",
		Description: "Add subproject tracker task support",
		Migrate: func(tx *xorm.Engine) error {
			exists, err := columnExists(tx, "tasks", "tracked_project_id")
			if err != nil {
				return err
			}
			if !exists {
				if _, err = tx.Exec("ALTER TABLE tasks ADD COLUMN tracked_project_id bigint NULL"); err != nil {
					return err
				}
			}

			var indexQuery string
			switch db.Type() {
			case schemas.POSTGRES, schemas.SQLITE:
				indexQuery = "CREATE INDEX IF NOT EXISTS IDX_tasks_tracked_project_id ON tasks (tracked_project_id)"
			case schemas.MYSQL:
				indexQuery = "CREATE INDEX IDX_tasks_tracked_project_id ON tasks (tracked_project_id)"
			}
			if _, err = tx.Exec(indexQuery); err != nil {
				// For MySQL, ignore duplicate key name error (Error 1061)
				if !strings.Contains(err.Error(), "Error 1061") && !strings.Contains(err.Error(), "Duplicate key name") {
					return err
				}
			}

			exists, err = columnExists(tx, "users", "subproject_task_title_template")
			if err != nil {
				return err
			}
			if !exists {
				if _, err = tx.Exec("ALTER TABLE users ADD COLUMN subproject_task_title_template varchar(250) NULL"); err != nil {
					return err
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
