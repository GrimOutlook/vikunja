<template>
	<Teleport to="body">
		<CustomTransition name="fade">
			<div
				v-if="isMenuOpen && targetTask"
				ref="menuRef"
				class="task-context-menu dropdown-menu show"
				:style="floatingStyle"
				tabindex="-1"
				data-cy="task-context-menu"
				@contextmenu.prevent
				@click.stop
			>
				<div class="dropdown-content">
					<!-- Main menu list -->
					<template v-if="activeSubPanel === 'none'">
						<!-- Action 1: Toggle Done / Undone -->
						<DropdownItem
							:icon="targetTask.done ? 'undo' : 'check'"
							data-cy="context-menu-mark-done"
							@click="toggleTaskDone"
						>
							{{ targetTask.done ? $t('task.actions.markUndone') : $t('task.actions.markDone') }}
						</DropdownItem>

						<!-- Action 2: Edit Title -->
						<DropdownItem
							icon="pen"
							data-cy="context-menu-edit-title"
							@click="toggleSubPanel('editTitle')"
						>
							{{ $t('task.actions.editTitle') }}
						</DropdownItem>

						<!-- Action 3: Due Date -->
						<DropdownItem
							icon="calendar"
							data-cy="context-menu-due-date"
							@click="toggleSubPanel('dueDate')"
						>
							{{ $t('task.attributes.dueDate') }}
						</DropdownItem>

						<!-- Action 4: Assignees -->
						<DropdownItem
							icon="user-plus"
							data-cy="context-menu-assignees"
							@click="toggleSubPanel('assignees')"
						>
							{{ $t('task.attributes.assignees') }}
						</DropdownItem>

						<!-- Action 5: Labels -->
						<DropdownItem
							icon="tag"
							data-cy="context-menu-labels"
							@click="toggleSubPanel('labels')"
						>
							{{ $t('task.attributes.labels') }}
						</DropdownItem>

						<!-- Action 6: Relations -->
						<DropdownItem
							icon="link"
							data-cy="context-menu-relations"
							@click="toggleSubPanel('relations')"
						>
							{{ $t('task.actions.relations') }}
						</DropdownItem>

						<hr class="dropdown-divider">

						<!-- Action 7: Delete Task -->
						<DropdownItem
							icon="trash"
							class="has-text-danger"
							data-cy="context-menu-delete"
							@click="deleteTask"
						>
							{{ $t('task.actions.delete') }}
						</DropdownItem>
					</template>

					<!-- Subpanel: Edit Title Inline Input -->
					<div
						v-else-if="activeSubPanel === 'editTitle'"
						class="subpanel edit-title-panel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('none')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<input
								ref="titleInputRef"
								v-model="editableTitle"
								class="input is-small"
								:placeholder="$t('task.attributes.title')"
								@keydown.enter.prevent="saveTitle"
								@keydown.esc.prevent="toggleSubPanel('none')"
							>
							<div class="subpanel-actions">
								<BaseButton
									class="is-small"
									@click="toggleSubPanel('none')"
								>
									{{ $t('misc.cancel') }}
								</BaseButton>
								<BaseButton
									class="is-small is-primary"
									:loading="isSavingTitle"
									@click="saveTitle"
								>
									{{ $t('misc.save') }}
								</BaseButton>
							</div>
						</div>
					</div>

					<!-- Subpanel: DatepickerInline -->
					<div
						v-else-if="activeSubPanel === 'dueDate'"
						class="subpanel datepicker-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('none')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<BaseButton
								v-if="targetTask.dueDate"
								class="is-small is-text has-text-danger"
								@click="clearDueDate"
							>
								{{ $t('task.detail.removeDueDate') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<DatepickerInline
								:model-value="targetTask.dueDate"
								:show-shortcuts="true"
								@update:modelValue="saveDueDate"
							/>
						</div>
					</div>

					<!-- Subpanel: EditAssignees -->
					<div
						v-else-if="activeSubPanel === 'assignees'"
						class="subpanel assignees-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('none')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<EditAssignees
								:model-value="targetTask.assignees"
								:task-id="targetTask.id"
								:project-id="targetTask.projectId"
								@update:modelValue="handleAssigneesUpdated"
							/>
						</div>
					</div>

					<!-- Subpanel: EditLabels -->
					<div
						v-else-if="activeSubPanel === 'labels'"
						class="subpanel labels-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('none')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<EditLabels
								:model-value="targetTask.labels"
								:task-id="targetTask.id"
								:creatable="true"
								@update:modelValue="handleLabelsUpdated"
							/>
						</div>
					</div>

					<!-- Subpanel: Relations Menu -->
					<div
						v-else-if="activeSubPanel === 'relations'"
						class="subpanel relations-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('none')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<DropdownItem
								icon="lock"
								data-cy="context-menu-link-blocker"
								@click="toggleSubPanel('linkBlocker')"
							>
								{{ $t('task.actions.linkBlocker') }}
							</DropdownItem>
							<DropdownItem
								icon="ban"
								data-cy="context-menu-link-blocked"
								@click="toggleSubPanel('linkBlockedTask')"
							>
								{{ $t('task.actions.linkBlockedTask') }}
							</DropdownItem>
						</div>
					</div>

					<!-- Subpanel: Link Blocker -->
					<div
						v-else-if="activeSubPanel === 'linkBlocker'"
						class="subpanel link-blocker-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('relations')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedRelationTask"
								:placeholder="$t('task.relation.searchPlaceholder')"
								:select-placeholder="''"
								:loading="isSearchingTasks"
								:search-results="foundRelationTasks"
								label="title"
								@search="findRelationTasks"
								@update:modelValue="(t) => saveRelation(t, RELATION_KIND.BLOCKED)"
							>
								<template #searchResult="{option: optTask}">
									<span
										v-if="typeof optTask !== 'string'"
										class="search-result"
										:class="{'is-strikethrough': optTask.done}"
									>
										{{ optTask.title }}
									</span>
									<span v-else>{{ optTask }}</span>
								</template>
							</Multiselect>
						</div>
					</div>

					<!-- Subpanel: Link Blocked Task -->
					<div
						v-else-if="activeSubPanel === 'linkBlockedTask'"
						class="subpanel link-blocked-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('relations')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedRelationTask"
								:placeholder="$t('task.relation.searchPlaceholder')"
								:select-placeholder="''"
								:loading="isSearchingTasks"
								:search-results="foundRelationTasks"
								label="title"
								@search="findRelationTasks"
								@update:modelValue="(t) => saveRelation(t, RELATION_KIND.BLOCKING)"
							>
								<template #searchResult="{option: optTask}">
									<span
										v-if="typeof optTask !== 'string'"
										class="search-result"
										:class="{'is-strikethrough': optTask.done}"
									>
										{{ optTask.title }}
									</span>
									<span v-else>{{ optTask }}</span>
								</template>
							</Multiselect>
						</div>
					</div>
				</div>
			</div>
		</CustomTransition>
	</Teleport>
</template>

<script setup lang="ts">
import {ref, computed, watch, nextTick, onUnmounted} from 'vue'
import {computePosition, flip, shift, autoUpdate, type VirtualElement} from '@floating-ui/dom'
import {onClickOutside} from '@vueuse/core'
import {useI18n} from 'vue-i18n'

import BaseButton from '@/components/base/BaseButton.vue'
import CustomTransition from '@/components/misc/CustomTransition.vue'
import DropdownItem from '@/components/misc/DropdownItem.vue'
import DatepickerInline from '@/components/input/DatepickerInline.vue'
import EditAssignees from '@/components/tasks/partials/EditAssignees.vue'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import Multiselect from '@/components/input/Multiselect.vue'

import {useTaskContextMenu} from '@/composables/useTaskContextMenu'
import {useTaskStore} from '@/stores/tasks'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {RELATION_KIND, type IRelationKind} from '@/types/IRelationKind'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'
import type {ILabel} from '@/modelTypes/ILabel'
import {error} from '@/message'

const props = withDefaults(
	defineProps<{
		task?: ITask | null
		x?: number
		y?: number
		isOpen?: boolean
	}>(),
	{
		task: null,
		x: undefined,
		y: undefined,
		isOpen: undefined,
	},
)

const emit = defineEmits<{
	(e: 'close'): void
	(e: 'taskUpdated', task: ITask): void
	(e: 'taskDeleted', taskId: number): void
}>()

const {t} = useI18n({useScope: 'global'})
const taskStore = useTaskStore()

const {
	isOpen: composableIsOpen,
	position: composablePosition,
	activeTask: composableActiveTask,
	closeContextMenu: composableClose,
} = useTaskContextMenu()

const isMenuOpen = computed(() => props.isOpen ?? composableIsOpen.value)
const targetTask = computed(() => props.task ?? composableActiveTask.value)
const menuPosition = computed(() => {
	if (props.x !== undefined && props.y !== undefined) {
		return {x: props.x, y: props.y}
	}
	return composablePosition.value
})

const menuRef = ref<HTMLElement | null>(null)
const floatingStyle = ref<Record<string, string>>({})
let cleanupFloating: (() => void) | null = null

type SubPanelType = 'none' | 'editTitle' | 'dueDate' | 'assignees' | 'labels' | 'relations' | 'linkBlocker' | 'linkBlockedTask'
const activeSubPanel = ref<SubPanelType>('none')

const editableTitle = ref('')
const isSavingTitle = ref(false)
const titleInputRef = ref<HTMLInputElement | null>(null)

const selectedRelationTask = ref<ITask | null>(null)
const foundRelationTasks = ref<ITask[]>([])
const isSearchingTasks = ref(false)

// Virtual Element position for floating-ui
const virtualElement = computed<VirtualElement>(() => ({
	getBoundingClientRect(): DOMRect {
		const x = menuPosition.value.x ?? 0
		const y = menuPosition.value.y ?? 0
		return {
			width: 0,
			height: 0,
			x,
			y,
			top: y,
			left: x,
			right: x,
			bottom: y,
			toJSON: () => {},
		}
	},
}))

async function updatePosition() {
	if (!menuRef.value || !isMenuOpen.value) return

	await nextTick()

	const {x, y} = await computePosition(virtualElement.value, menuRef.value, {
		strategy: 'fixed',
		placement: 'bottom-start',
		middleware: [
			flip({fallbackPlacements: ['bottom-end', 'top-start', 'top-end', 'right-start', 'left-start']}),
			shift({padding: 8}),
		],
	})

	floatingStyle.value = {
		position: 'fixed',
		left: `${x}px`,
		top: `${y}px`,
		zIndex: '1000',
	}
}

function setupFloating() {
	stopFloating()
	if (menuRef.value && isMenuOpen.value) {
		updatePosition()
		cleanupFloating = autoUpdate(virtualElement.value, menuRef.value, updatePosition)
	}
}

function stopFloating() {
	if (cleanupFloating) {
		cleanupFloating()
		cleanupFloating = null
	}
}

function close() {
	if (props.isOpen !== undefined) {
		emit('close')
	}
	composableClose()
}

// Click outside handling
onClickOutside(menuRef, () => {
	if (isMenuOpen.value) {
		close()
	}
})

// Keyboard handling for Escape
function handleKeyDown(e: KeyboardEvent) {
	if (e.key === 'Escape' && isMenuOpen.value) {
		close()
	}
}

watch(
	() => [isMenuOpen.value, menuPosition.value.x, menuPosition.value.y],
	async ([open]) => {
		if (open) {
			activeSubPanel.value = 'none'
			if (targetTask.value) {
				editableTitle.value = targetTask.value.title
			}
			await nextTick()
			setupFloating()
			window.addEventListener('keydown', handleKeyDown)
		} else {
			stopFloating()
			window.removeEventListener('keydown', handleKeyDown)
		}
	},
	{immediate: true},
)

onUnmounted(() => {
	stopFloating()
	window.removeEventListener('keydown', handleKeyDown)
})

function toggleSubPanel(panel: SubPanelType) {
	if (activeSubPanel.value === panel) {
		activeSubPanel.value = 'none'
	} else {
		activeSubPanel.value = panel
		if (panel === 'editTitle') {
			nextTick(() => {
				titleInputRef.value?.focus()
			})
		} else if (panel === 'relations' || panel === 'linkBlocker' || panel === 'linkBlockedTask') {
			ensureTargetTaskRelations()
		}
	}
	nextTick(() => {
		updatePosition()
	})
}

// Action Handlers

// 1. Mark Done / Undone
async function toggleTaskDone() {
	if (!targetTask.value) return
	try {
		const updated = await taskStore.update({
			...targetTask.value,
			done: !targetTask.value.done,
		})
		emit('taskUpdated', updated)
		close()
	} catch (e) {
		console.error('Failed to toggle task done state', e)
	}
}

// 2. Edit Title
async function saveTitle() {
	if (!targetTask.value) return
	const trimmed = editableTitle.value.trim()
	if (trimmed === '') {
		error({message: t('task.detail.titleRequired')})
		return
	}
	if (trimmed === targetTask.value.title) {
		activeSubPanel.value = 'none'
		return
	}

	try {
		isSavingTitle.value = true
		const updated = await taskStore.update({
			...targetTask.value,
			title: trimmed,
		})
		emit('taskUpdated', updated)
		activeSubPanel.value = 'none'
		close()
	} catch (e) {
		console.error('Failed to save task title', e)
	} finally {
		isSavingTitle.value = false
	}
}

// 3. Due Date
async function saveDueDate(newDate: Date | null) {
	if (!targetTask.value) return
	try {
		const updated = await taskStore.update({
			...targetTask.value,
			dueDate: newDate,
		})
		emit('taskUpdated', updated)
		activeSubPanel.value = 'none'
		close()
	} catch (e) {
		console.error('Failed to save due date', e)
	}
}

async function clearDueDate() {
	await saveDueDate(null)
}

// 4. Assignees Updated
function handleAssigneesUpdated(newAssignees: IUser[] | undefined) {
	if (!targetTask.value) return
	targetTask.value.assignees = newAssignees ?? []
	emit('taskUpdated', targetTask.value)
}

// 5. Labels Updated
function handleLabelsUpdated(newLabels: ILabel[]) {
	if (!targetTask.value) return
	targetTask.value.labels = newLabels
	emit('taskUpdated', targetTask.value)
}

// 6. Delete Task
async function deleteTask() {
	if (!targetTask.value) return
	try {
		const taskId = targetTask.value.id
		await taskStore.delete(targetTask.value)
		emit('taskDeleted', taskId)
		close()
	} catch (e) {
		console.error('Failed to delete task', e)
	}
}

// 7. Relation Handlers
function getExistingRelatedTaskIds(task: ITask | null): Set<number> {
	const ids = new Set<number>()
	if (!task) return ids
	if (task.id) {
		ids.add(Number(task.id))
	}
	if (task.relatedTasks) {
		Object.values(task.relatedTasks).forEach(taskList => {
			if (Array.isArray(taskList)) {
				taskList.forEach(relTask => {
					if (relTask && relTask.id) {
						ids.add(Number(relTask.id))
					}
				})
			}
		})
	}
	return ids
}

async function ensureTargetTaskRelations() {
	if (!targetTask.value || !targetTask.value.id) return
	try {
		const taskService = new TaskService()
		const fullTask = await taskService.get(new TaskModel({id: targetTask.value.id})) as ITask
		if (fullTask && fullTask.relatedTasks) {
			targetTask.value.relatedTasks = fullTask.relatedTasks
		}
	} catch (e) {
		console.error('Failed to load task relations', e)
	}
}

async function findRelationTasks(query: string) {
	if (!query || query.trim() === '') {
		foundRelationTasks.value = []
		return
	}
	try {
		isSearchingTasks.value = true
		const taskService = new TaskService()
		const result = await taskService.getAll({}, {
			s: query,
			sort_by: 'done',
		}) as ITask[]
		const existingIds = getExistingRelatedTaskIds(targetTask.value)
		foundRelationTasks.value = result.filter(t => !existingIds.has(Number(t.id)))
	} catch (e) {
		console.error('Failed to search tasks for relation', e)
	} finally {
		isSearchingTasks.value = false
	}
}

async function saveRelation(otherTask: ITask | null | string, kind: IRelationKind) {
	if (!targetTask.value || !otherTask || typeof otherTask === 'string' || !otherTask.id) return
	try {
		const taskRelationService = new TaskRelationService()
		await taskRelationService.create(new TaskRelationModel({
			taskId: targetTask.value.id,
			otherTaskId: otherTask.id,
			relationKind: kind,
		}))
		if (!targetTask.value.relatedTasks) {
			targetTask.value.relatedTasks = {}
		}
		if (!targetTask.value.relatedTasks[kind]) {
			targetTask.value.relatedTasks[kind] = []
		}
		targetTask.value.relatedTasks[kind]!.push(otherTask)
		const updated = await taskStore.update(targetTask.value)
		emit('taskUpdated', updated)
		selectedRelationTask.value = null
		foundRelationTasks.value = []
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to create relation', e)
	}
}
</script>

<style scoped lang="scss">
.task-context-menu {
	display: block;
	min-inline-size: 200px;
	max-inline-size: 320px;
	padding: 0.25rem 0;
	box-shadow: 0 0.5em 1em -0.125em rgba(10, 10, 10, 0.15), 0 0 0 1px rgba(10, 10, 10, 0.05);
	border-radius: $radius;
	background-color: var(--scheme-main);
	color: var(--text);
	border: 1px solid var(--grey-200);

	:deep(.search-result-button) {
		color: var(--text);

		&:hover,
		&:focus {
			background: var(--grey-100);
			color: var(--text);
		}
	}

	.dropdown-content {
		padding: 0.25rem;
		box-shadow: none;
	}

	.subpanel {
		padding: 0.5rem;

		.subpanel-header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			margin-block-end: 0.5rem;
			border-block-end: 1px solid var(--grey-100);
			padding-block-end: 0.25rem;
		}

		.subpanel-body {
			display: flex;
			flex-direction: column;
			gap: 0.5rem;
		}

		.subpanel-actions {
			display: flex;
			justify-content: flex-end;
			gap: 0.5rem;
		}
	}
}
</style>
