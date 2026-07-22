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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// task20260721180000 mirrors just the tasks columns this migration reads or
// writes: the tracker's own row (id, percent_done, the new count columns) and,
// via the same struct, the tracked subproject's tasks it counts (project_id).
type task20260721180000 struct {
	ID                       int64 `xorm:"pk autoincr"`
	ProjectID                int64
	TrackedProjectID         *int64
	PercentDone              float64
	SubprojectDoneTaskCount  *int64
	SubprojectTotalTaskCount *int64
}

func (task20260721180000) TableName() string {
	return "tasks"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260721180000",
		Description: "Add subproject tracker done/total task counts",
		Migrate: func(tx *xorm.Engine) error {
			columns := []string{"subproject_done_task_count", "subproject_total_task_count"}
			for _, col := range columns {
				exists, err := columnExists(tx, "tasks", col)
				if err != nil {
					return err
				}
				if exists {
					continue
				}

				if _, err = tx.Exec("ALTER TABLE tasks ADD COLUMN " + col + " bigint NULL"); err != nil {
					return err
				}
			}

			// Backfill: without this, tracker tasks created before this migration
			// keep NULL counts until something in their subproject changes again.
			trackers := []*task20260721180000{}
			if err := tx.Where("tracked_project_id IS NOT NULL").Find(&trackers); err != nil {
				return err
			}

			for _, tracker := range trackers {
				total, err := tx.
					Where("project_id = ? AND deleted_at IS NULL", *tracker.TrackedProjectID).
					Count(&task20260721180000{})
				if err != nil {
					return err
				}

				done, err := tx.
					Where("project_id = ? AND deleted_at IS NULL AND done = ?", *tracker.TrackedProjectID, true).
					Count(&task20260721180000{})
				if err != nil {
					return err
				}

				var percentDone float64
				if total > 0 {
					percentDone = float64(done) / float64(total)
				}

				_, err = tx.ID(tracker.ID).
					Cols("percent_done", "subproject_done_task_count", "subproject_total_task_count").
					Update(&task20260721180000{
						PercentDone:              percentDone,
						SubprojectDoneTaskCount:  &done,
						SubprojectTotalTaskCount: &total,
					})
				if err != nil {
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
