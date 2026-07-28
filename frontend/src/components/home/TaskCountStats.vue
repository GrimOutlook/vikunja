<script lang="ts" setup>
import {ref, onMounted} from 'vue'

import TaskService from '@/services/task'

const openCount = ref<number | null>(null)
const doneCount = ref<number | null>(null)

onMounted(async () => {
	const taskService = new TaskService()

	try {
		const [open, done] = await Promise.all([
			taskService.countAll('done = false'),
			taskService.countAll('done = true'),
		])
		openCount.value = open
		doneCount.value = done
	} catch {
		// These counts are supplementary, so a failure hides them instead of
		// interrupting the overview with an error.
	}
})
</script>

<template>
	<div
		v-if="openCount !== null && doneCount !== null"
		class="task-count-stats"
	>
		<div class="task-count-stats__stat">
			<span
				v-cy="'taskCountOpen'"
				class="task-count-stats__value"
			>{{ openCount }}</span>
			<span class="task-count-stats__label">{{ $t('home.taskCount.open') }}</span>
		</div>
		<div class="task-count-stats__stat">
			<span
				v-cy="'taskCountDone'"
				class="task-count-stats__value"
			>{{ doneCount }}</span>
			<span class="task-count-stats__label">{{ $t('home.taskCount.done') }}</span>
		</div>
	</div>
</template>

<style lang="scss" scoped>
.task-count-stats {
	display: flex;
	justify-content: center;
	gap: 3rem;
}

.task-count-stats__stat {
	display: flex;
	flex-direction: column;
}

.task-count-stats__value {
	font-size: 1.75rem;
	font-weight: bold;
	line-height: 1.2;
	color: var(--text);
}

.task-count-stats__label {
	font-size: .9rem;
	color: var(--grey-500);
}
</style>
