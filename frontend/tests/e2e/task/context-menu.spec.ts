import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {createDefaultViews} from '../project/prepareProjects'

async function setupProjectWithViewsAndTasks() {
	const projects = await ProjectFactory.create(1)
	const project = projects[0]

	// Create 4 views: List (1), Gantt (2), Table (3), Kanban (4)
	await createDefaultViews(project.id, 1)

	// Create Kanban bucket for Kanban view (view_kind: 3, id: 4)
	const buckets = await BucketFactory.create(1, {
		project_view_id: 4,
	})

	// Create test task with start/due dates so it renders across all views (List, Kanban, Gantt, Table)
	const now = new Date()
	const tomorrow = new Date(now.getTime() + 86400000)
	const tasks = await TaskFactory.create(1, {
		project_id: project.id,
		title: 'Context Menu Test Task',
		start_date: now.toISOString(),
		end_date: tomorrow.toISOString(),
		due_date: tomorrow.toISOString(),
	})

	// Associate task with Kanban bucket
	await TaskBucketFactory.create(1, {
		task_id: tasks[0].id,
		bucket_id: buckets[0].id,
		project_view_id: 4,
	}, false)

	return {
		project,
		task: tasks[0],
		bucket: buckets[0],
	}
}

test.describe('Task Context Menu', () => {
	test('Opens context menu on right click in List view and displays action options', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		await page.goto(`/projects/${project.id}/1`)

		const taskItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(taskItem).toBeVisible()

		await taskItem.click({button: 'right'})

		const contextMenu = page.locator('[data-cy="task-context-menu"]')
		await expect(contextMenu).toBeVisible()

		await expect(contextMenu.locator('[data-cy="context-menu-mark-done"]')).toBeVisible()
		await expect(contextMenu.locator('[data-cy="context-menu-edit-title"]')).toBeVisible()
		await expect(contextMenu.locator('[data-cy="context-menu-due-date"]')).toBeVisible()
		await expect(contextMenu.locator('[data-cy="context-menu-assignees"]')).toBeVisible()
		await expect(contextMenu.locator('[data-cy="context-menu-labels"]')).toBeVisible()
		await expect(contextMenu.locator('[data-cy="context-menu-delete"]')).toBeVisible()
	})

	test('Marks task as done and undone via context menu', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		await page.goto(`/projects/${project.id}/1`)

		const taskItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(taskItem).toBeVisible()

		// Right-click task and mark as done
		await taskItem.click({button: 'right'})
		const contextMenu = page.locator('[data-cy="task-context-menu"]')
		await expect(contextMenu).toBeVisible()
		await contextMenu.locator('[data-cy="context-menu-mark-done"]').click()

		// Context menu should close and task should reflect done state (.tasktext.done)
		await expect(contextMenu).not.toBeVisible()
		await expect(taskItem.locator('.tasktext')).toHaveClass(/done/)

		// Right-click task and mark as undone
		await taskItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await contextMenu.locator('[data-cy="context-menu-mark-done"]').click()

		// Context menu should close and task should reflect undone state
		await expect(contextMenu).not.toBeVisible()
		await expect(taskItem.locator('.tasktext')).not.toHaveClass(/done/)
	})

	test('Edits task title via context menu inline input', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		await page.goto(`/projects/${project.id}/1`)

		const taskItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(taskItem).toBeVisible()

		await taskItem.click({button: 'right'})
		const contextMenu = page.locator('[data-cy="task-context-menu"]')
		await expect(contextMenu).toBeVisible()

		// Click Edit Title
		await contextMenu.locator('[data-cy="context-menu-edit-title"]').click()

		// Subpanel input should appear
		const input = contextMenu.locator('input.input')
		await expect(input).toBeVisible()

		const updatedTitle = 'Renamed Task Title via Context Menu'
		await input.fill(updatedTitle)
		await input.press('Enter')

		// Context menu should close and UI updated with new task title
		await expect(contextMenu).not.toBeVisible()
		await expect(page.locator('.tasks .task')).toContainText(updatedTitle)
	})

	test('Deletes task via context menu', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		await page.goto(`/projects/${project.id}/1`)

		const taskItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(taskItem).toBeVisible()

		await taskItem.click({button: 'right'})
		const contextMenu = page.locator('[data-cy="task-context-menu"]')
		await expect(contextMenu).toBeVisible()

		// Click Delete, then confirm - the menu asks in a subpanel of its own
		await contextMenu.locator('[data-cy="context-menu-delete"]').click()
		await contextMenu.locator('[data-cy="context-menu-confirm-delete"]').click()

		// Context menu should close and task item removed from list
		await expect(contextMenu).not.toBeVisible()
		await expect(page.locator('.tasks .task').filter({hasText: task.title})).not.toBeVisible()
	})

	test('Opens context menu across all 4 project views (List, Kanban, Gantt, Table)', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		const contextMenu = page.locator('[data-cy="task-context-menu"]')

		// 1. List View
		await page.goto(`/projects/${project.id}/1`)
		const listItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(listItem).toBeVisible()
		await listItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await page.keyboard.press('Escape')
		await expect(contextMenu).not.toBeVisible()

		// 2. Kanban View
		await page.goto(`/projects/${project.id}/4`)
		const kanbanItem = page.locator('.kanban .bucket .tasks .task').filter({hasText: task.title}).first()
		await expect(kanbanItem).toBeVisible()
		await kanbanItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await page.keyboard.press('Escape')
		await expect(contextMenu).not.toBeVisible()

		// 3. Gantt View
		await page.goto(`/projects/${project.id}/2`)
		const ganttItem = page.locator('.gantt-bar').first()
		await expect(ganttItem).toBeVisible({timeout: 10000})
		await ganttItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await page.keyboard.press('Escape')
		await expect(contextMenu).not.toBeVisible()

		// 4. Table View
		await page.goto(`/projects/${project.id}/3`)
		const tableRow = page.locator('.project-table table.table tbody tr').filter({hasText: task.title}).first()
		await expect(tableRow).toBeVisible()
		await tableRow.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await page.keyboard.press('Escape')
		await expect(contextMenu).not.toBeVisible()
	})

	test('Opens subpanels for Due Date, Assignees, and Labels', async ({authenticatedPage: page}) => {
		const {project, task} = await setupProjectWithViewsAndTasks()
		await page.goto(`/projects/${project.id}/1`)

		const taskItem = page.locator('.tasks .task').filter({hasText: task.title}).first()
		await expect(taskItem).toBeVisible()

		// Check Due Date subpanel
		await taskItem.click({button: 'right'})
		const contextMenu = page.locator('[data-cy="task-context-menu"]')
		await expect(contextMenu).toBeVisible()
		await contextMenu.locator('[data-cy="context-menu-due-date"]').click()
		await expect(contextMenu.locator('.datepicker-subpanel')).toBeVisible()
		await page.keyboard.press('Escape')

		// Check Assignees subpanel
		await taskItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await contextMenu.locator('[data-cy="context-menu-assignees"]').click()
		await expect(contextMenu.locator('.assignees-subpanel')).toBeVisible()
		await page.keyboard.press('Escape')

		// Check Labels subpanel
		await taskItem.click({button: 'right'})
		await expect(contextMenu).toBeVisible()
		await contextMenu.locator('[data-cy="context-menu-labels"]').click()
		await expect(contextMenu.locator('.labels-subpanel')).toBeVisible()
		await page.keyboard.press('Escape')
	})
})
