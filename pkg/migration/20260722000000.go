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
	"time"

	"github.com/google/uuid"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

// The tracker task title placeholder/default template, mirrored here rather
// than imported from pkg/models: migrations stay independent of model code
// that's free to keep evolving after this migration is written.
const (
	subprojectTrackerTitlePlaceholder20260722000000     = "{{Subproject}}"
	defaultSubprojectTrackerTitleTemplate20260722000000 = "{{Subproject}} Subproject"
)

type project20260722000000 struct {
	ID              int64 `xorm:"pk autoincr"`
	Title           string
	OwnerID         int64
	ParentProjectID *int64
}

func (project20260722000000) TableName() string {
	return "projects"
}

type projectView20260722000000 struct {
	ID        int64 `xorm:"pk autoincr"`
	ProjectID int64
}

func (projectView20260722000000) TableName() string {
	return "project_views"
}

type user20260722000000 struct {
	ID                          int64 `xorm:"pk autoincr"`
	SubprojectTaskTitleTemplate string
}

func (user20260722000000) TableName() string {
	return "users"
}

// task20260722000000 mirrors just the tasks columns a newly-created tracker
// task needs. Created/Updated keep the same xorm tags as the real Task
// struct so xorm auto-populates them on Insert the same way; every other
// NOT NULL column not listed here (repeat_mode, cover_image_attachment_id,
// ...) has a DB-level default from the migration that first created it.
type task20260722000000 struct {
	ID                       int64 `xorm:"pk autoincr"`
	Title                    string
	ProjectID                int64
	TrackedProjectID         *int64
	CreatedByID              int64
	UID                      string
	Index                    int64 `xorm:"'index'"`
	PercentDone              float64
	SubprojectDoneTaskCount  *int64
	SubprojectTotalTaskCount *int64
	Done                     bool
	Created                  time.Time `xorm:"created"`
	Updated                  time.Time `xorm:"updated"`
}

func (task20260722000000) TableName() string {
	return "tasks"
}

type taskPosition20260722000000 struct {
	TaskID        int64 `xorm:"pk"`
	ProjectViewID int64 `xorm:"pk"`
	Position      float64
}

func (taskPosition20260722000000) TableName() string {
	return "task_positions"
}

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260722000000",
		Description: "Create subproject tracker tasks for subprojects that predate the feature",
		Migrate: func(tx *xorm.Engine) error {
			// Tracker task creation is hooked into project creation
			// (models.Project.Create), so it only ever ran for subprojects
			// created after this feature shipped. Subprojects from before
			// that never got one and need it backfilled here.
			subprojects := []*project20260722000000{}
			if err := tx.
				Where("parent_project_id IS NOT NULL AND parent_project_id != 0").
				Find(&subprojects); err != nil {
				return err
			}

			for _, subproject := range subprojects {
				has, err := tx.Where("tracked_project_id = ?", subproject.ID).Exist(&task20260722000000{})
				if err != nil {
					return err
				}
				if has {
					continue
				}

				owner := &user20260722000000{}
				if _, err = tx.ID(subproject.OwnerID).Get(owner); err != nil {
					return err
				}

				template := owner.SubprojectTaskTitleTemplate
				if template == "" {
					template = defaultSubprojectTrackerTitleTemplate20260722000000
				}
				title := strings.ReplaceAll(template, subprojectTrackerTitlePlaceholder20260722000000, subproject.Title)

				var maxIndex int64
				if _, err = tx.Table("tasks").
					Where("project_id = ?", *subproject.ParentProjectID).
					Select("COALESCE(MAX(`index`), 0)").
					Get(&maxIndex); err != nil {
					return err
				}
				index := maxIndex + 1

				total, err := tx.Where("project_id = ?", subproject.ID).Count(&task20260722000000{})
				if err != nil {
					return err
				}
				done, err := tx.Where("project_id = ? AND done = ?", subproject.ID, true).Count(&task20260722000000{})
				if err != nil {
					return err
				}
				var percentDone float64
				if total > 0 {
					percentDone = float64(done) / float64(total)
				}

				subprojectID := subproject.ID
				tracker := &task20260722000000{
					Title:                    title,
					ProjectID:                *subproject.ParentProjectID,
					TrackedProjectID:         &subprojectID,
					CreatedByID:              subproject.OwnerID,
					UID:                      uuid.NewString(),
					Index:                    index,
					PercentDone:              percentDone,
					SubprojectDoneTaskCount:  &done,
					SubprojectTotalTaskCount: &total,
				}
				if _, err = tx.Insert(tracker); err != nil {
					return err
				}

				// Position the tracker task in every view of the parent
				// project, same as a normal task creation would, just with
				// a simplified (index-based) default position rather than
				// replicating the full "insert at the top" recalculation
				// logic - correct ordering here is cosmetic, not required
				// for the task to show up.
				views := []*projectView20260722000000{}
				if err = tx.Where("project_id = ?", *subproject.ParentProjectID).Find(&views); err != nil {
					return err
				}
				for _, view := range views {
					position := &taskPosition20260722000000{
						TaskID:        tracker.ID,
						ProjectViewID: view.ID,
						Position:      float64(index) * 65536,
					}
					if _, err = tx.Insert(position); err != nil {
						return err
					}
				}
			}

			return nil
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
