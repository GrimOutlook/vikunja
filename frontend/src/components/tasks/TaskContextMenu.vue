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
							v-if="!isMulti"
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

					<!-- Subpanel: Assignees Menu -->
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
							<DropdownItem
								icon="user-plus"
								data-cy="context-menu-add-assignee"
								@click="openAssigneeSubpanel('addAssignee')"
							>
								{{ $t('task.actions.addAssignee') }}
							</DropdownItem>
							<DropdownItem
								icon="user-minus"
								data-cy="context-menu-remove-assignee"
								@click="openAssigneeSubpanel('removeAssignee')"
							>
								{{ $t('task.actions.removeAssignee') }}
							</DropdownItem>
						</div>
					</div>

					<!-- Subpanel: Add Assignee -->
					<div
						v-else-if="activeSubPanel === 'addAssignee'"
						class="subpanel add-assignee-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('assignees')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<span class="subpanel-title">{{ $t('task.actions.addAssignee') }}</span>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedAssignee"
								:placeholder="$t('task.assignee.placeholder')"
								:loading="isSearchingUsers"
								:search-results="foundUsers"
								label="name"
								:select-placeholder="''"
								@search="findUsers"
								@update:modelValue="handleAddAssignee"
							>
								<template #searchResult="{option: user}">
									<User
										v-if="typeof user !== 'string'"
										:avatar-size="24"
										:show-username="true"
										:user="user"
									/>
								</template>
							</Multiselect>
						</div>
					</div>

					<!-- Subpanel: Remove Assignee -->
					<div
						v-else-if="activeSubPanel === 'removeAssignee'"
						class="subpanel remove-assignee-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('assignees')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<span class="subpanel-title">{{ $t('task.actions.removeAssignee') }}</span>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedAssignee"
								:placeholder="$t('task.assignee.placeholder')"
								:loading="isSearchingUsers"
								:search-results="foundUsers"
								label="name"
								:select-placeholder="''"
								@search="findUsers"
								@update:modelValue="handleRemoveAssignee"
							>
								<template #searchResult="{option: user}">
									<User
										v-if="typeof user !== 'string'"
										:avatar-size="24"
										:show-username="true"
										:user="user"
									/>
								</template>
							</Multiselect>
						</div>
					</div>

					<!-- Subpanel: Labels Menu -->
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
							<DropdownItem
								icon="tag"
								data-cy="context-menu-add-label"
								@click="openLabelSubpanel('addLabel')"
							>
								{{ $t('task.actions.addLabel') }}
							</DropdownItem>
							<DropdownItem
								icon="tag"
								data-cy="context-menu-remove-label"
								@click="openLabelSubpanel('removeLabel')"
							>
								{{ $t('task.actions.removeLabel') }}
							</DropdownItem>
						</div>
					</div>

					<!-- Subpanel: Add Label -->
					<div
						v-else-if="activeSubPanel === 'addLabel'"
						class="subpanel add-label-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('labels')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<span class="subpanel-title">{{ $t('task.actions.addLabel') }}</span>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedLabel"
								:placeholder="$t('task.label.placeholder')"
								:loading="labelStore.isLoading"
								:search-results="foundLabels"
								label="title"
								:select-placeholder="''"
								@search="findLabels"
								@update:modelValue="handleAddLabel"
							>
								<template #searchResult="{option: label}">
									<span
										v-if="typeof label !== 'string'"
										:style="getLabelStyles(label)"
										class="tag search-result"
									>
										<span>{{ label.title }}</span>
									</span>
								</template>
							</Multiselect>
						</div>
					</div>

					<!-- Subpanel: Remove Label -->
					<div
						v-else-if="activeSubPanel === 'removeLabel'"
						class="subpanel remove-label-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('labels')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<span class="subpanel-title">{{ $t('task.actions.removeLabel') }}</span>
						</div>
						<div class="subpanel-body">
							<Multiselect
								v-model="selectedLabel"
								:placeholder="$t('task.label.placeholder')"
								:loading="labelStore.isLoading"
								:search-results="foundLabels"
								label="title"
								:select-placeholder="''"
								@search="findLabels"
								@update:modelValue="handleRemoveLabel"
							>
								<template #searchResult="{option: label}">
									<span
										v-if="typeof label !== 'string'"
										:style="getLabelStyles(label)"
										class="tag search-result"
									>
										<span>{{ label.title }}</span>
									</span>
								</template>
							</Multiselect>
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
								v-for="rk in RELATION_KINDS"
								:key="rk"
								:data-cy="`context-menu-relation-${rk}`"
								@click="openRelationTypeSubpanel(rk)"
							>
								{{ $t(`task.relation.kinds.${rk}`, 1) }}
							</DropdownItem>
						</div>
					</div>

					<!-- Subpanel: Add Relation -->
					<div
						v-else-if="activeSubPanel === 'addRelation' && selectedRelationKind"
						class="subpanel add-relation-subpanel"
					>
						<div class="subpanel-header">
							<BaseButton
								class="is-small is-ghost"
								@click="toggleSubPanel('relations')"
							>
								&larr; {{ $t('misc.back') }}
							</BaseButton>
							<span class="subpanel-title">{{ $t(`task.relation.kinds.${selectedRelationKind}`, 1) }}</span>
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
								@update:modelValue="(t) => saveRelation(t, selectedRelationKind!)"
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
import Multiselect from '@/components/input/Multiselect.vue'
import User from '@/components/misc/User.vue'

import {useTaskContextMenu} from '@/composables/useTaskContextMenu'
import {useTaskStore} from '@/stores/tasks'
import {useLabelStore} from '@/stores/labels'
import {useLabelStyles} from '@/composables/useLabelStyles'
import TaskService from '@/services/task'
import TaskModel from '@/models/task'
import ProjectUserService from '@/services/projectUsers'
import {getDisplayName} from '@/models/user'
import TaskRelationService from '@/services/taskRelation'
import TaskRelationModel from '@/models/taskRelation'
import {RELATION_KINDS, RELATION_KIND, type IRelationKind} from '@/types/IRelationKind'
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
const labelStore = useLabelStore()
const {getLabelStyles} = useLabelStyles()

const {
	isOpen: composableIsOpen,
	position: composablePosition,
	activeTask: composableActiveTask,
	closeContextMenu: composableClose,
	isMultiSelectMode,
	selectedTasks,
	clearTaskSelection,
} = useTaskContextMenu()

const isMenuOpen = computed(() => props.isOpen ?? composableIsOpen.value)
const targetTask = computed(() => props.task ?? composableActiveTask.value)

const selectedTaskItems = computed(() => {
	if (isMultiSelectMode.value && selectedTasks.value.length > 0) {
		return selectedTasks.value
	}
	if (targetTask.value) {
		return [targetTask.value]
	}
	return []
})

const isMulti = computed(() => selectedTaskItems.value.length > 1)

const menuPosition = computed(() => {
	if (props.x !== undefined && props.y !== undefined) {
		return {x: props.x, y: props.y}
	}
	return composablePosition.value
})

const menuRef = ref<HTMLElement | null>(null)
const floatingStyle = ref<Record<string, string>>({})
let cleanupFloating: (() => void) | null = null

type SubPanelType = 'none' | 'editTitle' | 'dueDate' | 'assignees' | 'addAssignee' | 'removeAssignee' | 'labels' | 'addLabel' | 'removeLabel' | 'relations' | 'addRelation' | 'linkBlocker' | 'linkBlockedTask'
const activeSubPanel = ref<SubPanelType>('none')

const editableTitle = ref('')
const isSavingTitle = ref(false)
const titleInputRef = ref<HTMLInputElement | null>(null)

const selectedAssignee = ref<IUser | null>(null)
const foundUsers = ref<IUser[]>([])
const isSearchingUsers = ref(false)

const selectedLabel = ref<ILabel | null>(null)
const labelQuery = ref('')

const foundLabels = computed(() => labelStore.filterLabelsByQuery([], labelQuery.value))

function openAssigneeSubpanel(panel: 'addAssignee' | 'removeAssignee') {
	selectedAssignee.value = null
	foundUsers.value = []
	findUsers('')
	toggleSubPanel(panel)
}

async function findUsers(query = '') {
	try {
		isSearchingUsers.value = true
		const projectUserService = new ProjectUserService()
		const projectId = targetTask.value?.projectId ?? 0
		const response = await projectUserService.getAll({projectId}, {s: query}) as IUser[]
		foundUsers.value = response.map(u => {
			u.name = getDisplayName(u)
			return u
		})
	} catch (e) {
		console.error('Failed to search users', e)
	} finally {
		isSearchingUsers.value = false
	}
}

async function handleAddAssignee(user: IUser | null | string) {
	if (!user || typeof user === 'string' || !user.id) return
	try {
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			await taskStore.addAssignee({user, taskId: t.id})
			emit('taskUpdated', t)
		}
		selectedAssignee.value = null
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to add assignee', e)
	}
}

async function handleRemoveAssignee(user: IUser | null | string) {
	if (!user || typeof user === 'string' || !user.id) return
	try {
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			await taskStore.removeAssignee({user, taskId: t.id})
			emit('taskUpdated', t)
		}
		selectedAssignee.value = null
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to remove assignee', e)
	}
}

function openLabelSubpanel(panel: 'addLabel' | 'removeLabel') {
	selectedLabel.value = null
	labelQuery.value = ''
	toggleSubPanel(panel)
}

function findLabels(query: string) {
	labelQuery.value = query
}

async function handleAddLabel(label: ILabel | null | string) {
	if (!label || typeof label === 'string' || !label.id) return
	try {
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			await taskStore.addLabel({label, taskId: t.id})
			emit('taskUpdated', t)
		}
		selectedLabel.value = null
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to add label', e)
	}
}

async function handleRemoveLabel(label: ILabel | null | string) {
	if (!label || typeof label === 'string' || !label.id) return
	try {
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			await taskStore.removeLabel({label, taskId: t.id})
			emit('taskUpdated', t)
		}
		selectedLabel.value = null
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to remove label', e)
	}
}

const selectedRelationKind = ref<IRelationKind | null>(null)
const selectedRelationTask = ref<ITask | null>(null)
const foundRelationTasks = ref<ITask[]>([])
const isSearchingTasks = ref(false)

function openRelationTypeSubpanel(kind: IRelationKind) {
	selectedRelationKind.value = kind
	selectedRelationTask.value = null
	foundRelationTasks.value = []
	toggleSubPanel('addRelation')
}

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
		} else if (panel === 'relations' || panel === 'addRelation' || panel === 'linkBlocker' || panel === 'linkBlockedTask') {
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
	if (selectedTaskItems.value.length === 0) return
	try {
		const markDone = selectedTaskItems.value.some(t => !t.done)
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			const updated = await taskStore.update({
				...t,
				done: markDone,
			})
			emit('taskUpdated', updated)
		}
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
	if (selectedTaskItems.value.length === 0) return
	try {
		const tasksToUpdate = [...selectedTaskItems.value]
		for (const t of tasksToUpdate) {
			const updated = await taskStore.update({
				...t,
				dueDate: newDate,
			})
			emit('taskUpdated', updated)
		}
		toggleSubPanel('none')
		close()
	} catch (e) {
		console.error('Failed to save due date', e)
	}
}

async function clearDueDate() {
	await saveDueDate(null)
}

// 4. Delete Task
async function deleteTask() {
	if (selectedTaskItems.value.length === 0) return
	try {
		const tasksToDelete = [...selectedTaskItems.value]
		for (const t of tasksToDelete) {
			const taskId = t.id
			await taskStore.delete(t)
			emit('taskDeleted', taskId)
		}
		clearTaskSelection()
		close()
	} catch (e) {
		console.error('Failed to delete task', e)
	}
}

// 5. Relation Handlers
function getExistingRelatedTaskIds(tasks: ITask[]): Set<number> {
	const ids = new Set<number>()
	if (!tasks || tasks.length === 0) return ids

	tasks.forEach(task => {
		if (task && task.id) {
			ids.add(Number(task.id))
		}
		if (task && task.relatedTasks) {
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
	})

	return ids
}

async function ensureTargetTaskRelations() {
	if (selectedTaskItems.value.length === 0) return
	try {
		const taskService = new TaskService()
		for (const t of selectedTaskItems.value) {
			if (!t.id) continue
			try {
				const fullTask = await taskService.get(new TaskModel({id: t.id})) as ITask
				if (fullTask && fullTask.relatedTasks) {
					t.relatedTasks = fullTask.relatedTasks
				}
			} catch (e) {
				console.error(`Failed to load relations for task ${t.id}`, e)
			}
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
		const existingIds = getExistingRelatedTaskIds(selectedTaskItems.value)
		foundRelationTasks.value = result.filter(t => !existingIds.has(Number(t.id)))
	} catch (e) {
		console.error('Failed to search tasks for relation', e)
	} finally {
		isSearchingTasks.value = false
	}
}

async function saveRelation(otherTask: ITask | null | string, kind: IRelationKind) {
	if (!otherTask || typeof otherTask === 'string' || !otherTask.id) return
	const tasksToUpdate = [...selectedTaskItems.value]
	if (tasksToUpdate.length === 0) return

	try {
		const taskRelationService = new TaskRelationService()
		for (const t of tasksToUpdate) {
			try {
				await taskRelationService.create(new TaskRelationModel({
					taskId: t.id,
					otherTaskId: otherTask.id,
					relationKind: kind,
				}))
				if (!t.relatedTasks) {
					t.relatedTasks = {}
				}
				if (!t.relatedTasks[kind]) {
					t.relatedTasks[kind] = []
				}
				if (!t.relatedTasks[kind]!.some(rel => rel.id === (otherTask as ITask).id)) {
					t.relatedTasks[kind]!.push(otherTask as ITask)
				}
				const updated = await taskStore.update(t)
				emit('taskUpdated', updated)
			} catch (e) {
				console.error(`Failed to create relation for task ${t.id}`, e)
			}
		}
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
