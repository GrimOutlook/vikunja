import {ref} from 'vue'

import type {ITask} from '@/modelTypes/ITask'

interface ContextMenuPosition {
	x: number
	y: number
}

const isOpen = ref<boolean>(false)
const position = ref<ContextMenuPosition>({x: 0, y: 0})
const activeTask = ref<ITask | null>(null)

const isMultiSelectMode = ref<boolean>(true)
const selectedTasks = ref<ITask[]>([])

export function useTaskContextMenu() {
	function toggleTaskSelection(task: ITask) {
		const index = selectedTasks.value.findIndex(t => t.id === task.id)
		if (index !== -1) {
			selectedTasks.value.splice(index, 1)
		} else {
			selectedTasks.value.push(task)
		}
	}

	function selectAllTasks(tasksList: ITask[]) {
		selectedTasks.value = [...tasksList]
	}

	function clearTaskSelection() {
		selectedTasks.value = []
	}

	function isTaskSelected(taskId: number) {
		return selectedTasks.value.some(t => t.id === taskId)
	}

	function openContextMenu(e: MouseEvent, task: ITask) {
		e.preventDefault()
		e.stopPropagation()
		activeTask.value = task

		if (isMultiSelectMode.value && !isTaskSelected(task.id) && selectedTasks.value.length === 0) {
			selectedTasks.value = [task]
		}

		position.value = {x: e.clientX, y: e.clientY}
		isOpen.value = true
	}

	function closeContextMenu() {
		isOpen.value = false
		activeTask.value = null
	}

	return {
		isOpen,
		position,
		activeTask,
		isMultiSelectMode,
		selectedTasks,
		toggleTaskSelection,
		selectAllTasks,
		clearTaskSelection,
		isTaskSelected,
		openContextMenu,
		closeContextMenu,
	}
}
