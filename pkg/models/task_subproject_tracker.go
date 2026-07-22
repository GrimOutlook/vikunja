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

package models

import (
	"strings"

	"code.vikunja.io/api/pkg/user"
	"xorm.io/xorm"
)

// SubprojectTrackerTitlePlaceholder is replaced with the subproject's title when
// rendering the default title of a subproject tracker task.
const SubprojectTrackerTitlePlaceholder = "{{Subproject}}"

// DefaultSubprojectTrackerTitleTemplate is used when a user hasn't configured
// user.User.SubprojectTaskTitleTemplate.
const DefaultSubprojectTrackerTitleTemplate = "{{Subproject}} Subproject"

// renderSubprojectTrackerTitle fills the template with the subproject's title,
// falling back to the default template when none is configured.
func renderSubprojectTrackerTitle(template, subprojectTitle string) string {
	if template == "" {
		template = DefaultSubprojectTrackerTitleTemplate
	}
	return strings.ReplaceAll(template, SubprojectTrackerTitlePlaceholder, subprojectTitle)
}

// createSubprojectTrackerTask creates the system-managed task in subproject's parent
// project that tracks subproject's completion progress. Callers must already have
// verified subproject has a parent.
func createSubprojectTrackerTask(s *xorm.Session, subproject *Project, doer *user.User) error {
	subprojectID := subproject.ID

	tracker := &Task{
		Title:            renderSubprojectTrackerTitle(doer.SubprojectTaskTitleTemplate, subproject.Title),
		ProjectID:        subproject.parentID(),
		TrackedProjectID: &subprojectID,
	}

	if err := createTask(s, tracker, doer, false, false); err != nil {
		return err
	}

	// Populates the done/total counts as 0/0 immediately, rather than leaving
	// them unset until the subproject's first task creates a sync.
	return syncTrackerTaskForSubproject(s, subprojectID)
}

// deleteSubprojectTrackerTask permanently removes the tracker task for subprojectID,
// if one exists. It is a no-op if there is none (e.g. it was already removed as part
// of a parent project's own deletion cascade).
func deleteSubprojectTrackerTask(s *xorm.Session, subprojectID int64) error {
	tracker := &Task{}
	has, err := s.Where("tracked_project_id = ?", subprojectID).Get(tracker)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	return hardDeleteTask(s, tracker)
}

// calculateSubprojectProgress returns the number of subprojectID's non-deleted
// tasks that are done and the total count. Only tasks directly in the
// subproject are counted, not tasks in further-nested subprojects.
func calculateSubprojectProgress(s *xorm.Session, subprojectID int64) (done, total int64, err error) {
	total, err = s.Where("project_id = ?", subprojectID).Count(&Task{})
	if err != nil {
		return 0, 0, err
	}
	if total == 0 {
		return 0, 0, nil
	}

	done, err = s.Where("project_id = ? AND done = ?", subprojectID, true).Count(&Task{})
	if err != nil {
		return 0, 0, err
	}

	return done, total, nil
}

// syncTrackerTaskForSubproject recomputes and persists the percent_done and
// done/total task counts of subprojectID's tracker task, if one exists. It is
// a no-op if subprojectID has no parent project or its parent has no tracker
// task for it.
func syncTrackerTaskForSubproject(s *xorm.Session, subprojectID int64) error {
	subproject, err := GetProjectSimpleByID(s, subprojectID)
	if err != nil {
		if IsErrProjectDoesNotExist(err) {
			return nil
		}
		return err
	}

	if subproject.parentID() == 0 {
		return nil
	}

	tracker := &Task{}
	has, err := s.
		Where("project_id = ? AND tracked_project_id = ?", subproject.parentID(), subprojectID).
		Get(tracker)
	if err != nil {
		return err
	}
	if !has {
		return nil
	}

	done, total, err := calculateSubprojectProgress(s, subprojectID)
	if err != nil {
		return err
	}

	var percentDone float64
	if total > 0 {
		percentDone = float64(done) / float64(total)
	}

	_, err = s.ID(tracker.ID).
		Cols("percent_done", "subproject_done_task_count", "subproject_total_task_count").
		Update(&Task{
			PercentDone:              percentDone,
			SubprojectDoneTaskCount:  &done,
			SubprojectTotalTaskCount: &total,
		})
	return err
}
