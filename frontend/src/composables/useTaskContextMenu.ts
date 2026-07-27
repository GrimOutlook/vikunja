import {ref} from 'vue'

import type {ITask} from '@/modelTypes/ITask'

interface ContextMenuPosition {
	x: number
	y: number
}

const isOpen = ref<boolean>(false)
const position = ref<ContextMenuPosition>({x: 0, y: 0})
const activeTask = ref<ITask | null>(null)

export function useTaskContextMenu() {
	function openContextMenu(e: MouseEvent, task: ITask) {
		e.preventDefault()
		e.stopPropagation()
		activeTask.value = task
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
		openContextMenu,
		closeContextMenu,
	}
}
