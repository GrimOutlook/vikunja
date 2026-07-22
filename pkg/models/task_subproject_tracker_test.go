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
	"testing"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubprojectTracker(t *testing.T) {
	usr := &user.User{
		ID:       6,
		Username: "user6",
		Email:    "user6@example.com",
	}

	t.Run("creating a subproject creates a tracker task in the parent", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "tracker parent"}
		require.NoError(t, parent.Create(s, usr))

		sub := Project{Title: "tracker child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":         parent.ID,
			"tracked_project_id": sub.ID,
			"title":              "tracker child Subproject",
			"percent_done":       0,
		}, false)
	})

	t.Run("uses the user's configured title template", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		templatedUser := &user.User{
			ID:                          6,
			Username:                    "user6",
			Email:                       "user6@example.com",
			SubprojectTaskTitleTemplate: "{{Subproject}} Epic",
		}

		parent := Project{Title: "epic parent"}
		require.NoError(t, parent.Create(s, templatedUser))

		sub := Project{Title: "epic child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, templatedUser))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":         parent.ID,
			"tracked_project_id": sub.ID,
			"title":              "epic child Epic",
		}, false)
	})

	t.Run("top-level projects don't get a tracker task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		p := Project{Title: "no parent here"}
		require.NoError(t, p.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"tracked_project_id": p.ID,
		})
	})

	t.Run("deleting a subproject removes its tracker task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "delete tracker parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "delete tracker child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"tracked_project_id": sub.ID,
		}, false)

		s2 := db.NewSession()
		defer s2.Close()
		toDelete := Project{ID: sub.ID}
		require.NoError(t, toDelete.Delete(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"tracked_project_id": sub.ID,
		})
	})

	t.Run("deleting the parent project also removes the tracker task", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "cascade parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "cascade child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		s2 := db.NewSession()
		defer s2.Close()
		toDelete := Project{ID: parent.ID}
		require.NoError(t, toDelete.Delete(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertMissing(t, "tasks", map[string]interface{}{
			"tracked_project_id": sub.ID,
		})
		db.AssertMissing(t, "projects", map[string]interface{}{
			"id": sub.ID,
		})
	})

	t.Run("tracker percent_done follows the subproject's tasks", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "progress parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "progress child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))

		taskOne := Task{ProjectID: sub.ID, Title: "one"}
		require.NoError(t, taskOne.Create(s, usr))
		taskTwo := Task{ProjectID: sub.ID, Title: "two", Done: true}
		require.NoError(t, taskTwo.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":                  parent.ID,
			"tracked_project_id":          sub.ID,
			"percent_done":                0.5,
			"subproject_done_task_count":  1,
			"subproject_total_task_count": 2,
		}, false)

		s2 := db.NewSession()
		defer s2.Close()
		taskOne.Done = true
		require.NoError(t, taskOne.Update(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":                  parent.ID,
			"tracked_project_id":          sub.ID,
			"percent_done":                1,
			"subproject_done_task_count":  2,
			"subproject_total_task_count": 2,
		}, false)
	})

	t.Run("a freshly created subproject's tracker starts at 0/0, not unset", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "fresh parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "fresh child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":                  parent.ID,
			"tracked_project_id":          sub.ID,
			"percent_done":                0,
			"subproject_done_task_count":  0,
			"subproject_total_task_count": 0,
		}, false)
	})

	t.Run("deleting a subproject task recalculates the tracker's percent_done", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "delete-task parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "delete-task child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))

		taskOne := Task{ProjectID: sub.ID, Title: "one", Done: true}
		require.NoError(t, taskOne.Create(s, usr))
		taskTwo := Task{ProjectID: sub.ID, Title: "two"}
		require.NoError(t, taskTwo.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":         parent.ID,
			"tracked_project_id": sub.ID,
			"percent_done":       0.5,
		}, false)

		// Deleting the one not-done task should leave the subproject 100% done.
		s2 := db.NewSession()
		defer s2.Close()
		require.NoError(t, taskTwo.Delete(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"project_id":         parent.ID,
			"tracked_project_id": sub.ID,
			"percent_done":       1,
		}, false)
	})

	t.Run("moving a task between subprojects recalculates both trackers", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "move parent"}
		require.NoError(t, parent.Create(s, usr))
		subA := Project{Title: "move child a", ParentProjectID: &parent.ID}
		require.NoError(t, subA.Create(s, usr))
		subB := Project{Title: "move child b", ParentProjectID: &parent.ID}
		require.NoError(t, subB.Create(s, usr))

		// subA: 1 done task -> 100%. subB: empty -> 0%.
		taskInA := Task{ProjectID: subA.ID, Title: "in a", Done: true}
		require.NoError(t, taskInA.Create(s, usr))
		require.NoError(t, s.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"tracked_project_id": subA.ID,
			"percent_done":       1,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"tracked_project_id": subB.ID,
			"percent_done":       0,
		}, false)

		// Move the done task from subA to subB: subA becomes empty (0%),
		// subB gets its one done task (100%).
		s2 := db.NewSession()
		defer s2.Close()
		taskInA.ProjectID = subB.ID
		require.NoError(t, taskInA.Update(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"tracked_project_id": subA.ID,
			"percent_done":       0,
		}, false)
		db.AssertExists(t, "tasks", map[string]interface{}{
			"tracked_project_id": subB.ID,
			"percent_done":       1,
		}, false)
	})

	t.Run("a tracker task's percent_done can't be set directly", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "guard parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "guard child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		tracker := &Task{}
		has, err := s.Where("tracked_project_id = ?", sub.ID).Get(tracker)
		require.NoError(t, err)
		require.True(t, has)

		s2 := db.NewSession()
		defer s2.Close()
		tracker.PercentDone = 0.9
		require.NoError(t, tracker.Update(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":           tracker.ID,
			"percent_done": 0,
		}, false)
	})

	t.Run("a tracker task's title can be changed", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "title parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "title child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		tracker := &Task{}
		has, err := s.Where("tracked_project_id = ?", sub.ID).Get(tracker)
		require.NoError(t, err)
		require.True(t, has)

		s2 := db.NewSession()
		defer s2.Close()
		tracker.Title = "Renamed tracker"
		require.NoError(t, tracker.Update(s2, usr))
		require.NoError(t, s2.Commit())

		db.AssertExists(t, "tasks", map[string]interface{}{
			"id":    tracker.ID,
			"title": "Renamed tracker",
		}, false)
	})
}

func TestSubprojectTracker_Permissions(t *testing.T) {
	usr := &user.User{ID: 6, Username: "user6"}

	t.Run("cannot be created manually", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		trackedID := int64(42)
		task := &Task{ProjectID: 41, TrackedProjectID: &trackedID}
		can, err := task.CanCreate(s, usr)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot be deleted", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "no delete parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "no delete child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		require.NoError(t, s.Commit())

		tracker := &Task{}
		has, err := s.Where("tracked_project_id = ?", sub.ID).Get(tracker)
		require.NoError(t, err)
		require.True(t, has)

		can, err := tracker.CanDelete(s, usr)
		require.NoError(t, err)
		assert.False(t, can)
	})

	t.Run("cannot be moved to a different project", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		parent := Project{Title: "no move parent"}
		require.NoError(t, parent.Create(s, usr))
		sub := Project{Title: "no move child", ParentProjectID: &parent.ID}
		require.NoError(t, sub.Create(s, usr))
		other := Project{Title: "elsewhere"}
		require.NoError(t, other.Create(s, usr))
		require.NoError(t, s.Commit())

		tracker := &Task{}
		has, err := s.Where("tracked_project_id = ?", sub.ID).Get(tracker)
		require.NoError(t, err)
		require.True(t, has)

		s2 := db.NewSession()
		defer s2.Close()
		moved := &Task{ID: tracker.ID, ProjectID: other.ID}
		can, err := moved.CanUpdate(s2, usr)
		assert.False(t, can)
		assert.True(t, IsErrGenericForbidden(err))
	})
}
