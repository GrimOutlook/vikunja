<template>
	<Message
		v-if="task.trackedProjectId"
		variant="info"
		class="subproject-tracker-banner-wrapper details"
	>
		<div class="subproject-tracker-banner-text">
			<Icon icon="layer-group" />
			<span>{{ $t('task.detail.subprojectTracker.description') }}</span>
		</div>
		<div class="subproject-tracker-banner-progress">
			<ProgressBar
				:value="task.percentDone * 100"
				is-small
				is-primary
			/>
			<span>{{ Math.round(task.percentDone * 100) }}%</span>
			<span
				v-if="task.subprojectTotalTaskCount !== null"
				class="subproject-tracker-banner-count"
			>
				({{ task.subprojectDoneTaskCount }}/{{ task.subprojectTotalTaskCount }})
			</span>
		</div>
		<BaseButton
			class="subproject-tracker-banner-link"
			:to="{ name: 'project.index', params: { projectId: task.trackedProjectId } }"
		>
			<Icon icon="arrow-up-right-from-square" />
			{{ $t('task.detail.subprojectTracker.viewSubproject') }}
		</BaseButton>
	</Message>
</template>

<script lang="ts" setup>
import type {ITask} from '@/modelTypes/ITask'
import Message from '@/components/misc/Message.vue'
import ProgressBar from '@/components/misc/ProgressBar.vue'
import BaseButton from '@/components/base/BaseButton.vue'

defineProps<{
	task: ITask,
}>()
</script>

<style lang="scss" scoped>
// Message.vue paints its own solid background on this wrapper (its root
// element) underneath the inner .message div's translucent tinted one.
// Dropping it here avoids two stacked backgrounds show through at the edges.
.subproject-tracker-banner-wrapper {
	background: transparent;
}

// .message is a level deeper than Message's root, so reaching it needs :deep().
.subproject-tracker-banner-wrapper :deep(.message) {
	display: flex;
	flex-wrap: wrap;
	align-items: center;
	gap: .75rem;
	border: none;
}

.subproject-tracker-banner-text {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex: 1 1 auto;
}

.subproject-tracker-banner-progress {
	display: flex;
	align-items: center;
	gap: .5rem;
	flex: 0 0 auto;
}

.subproject-tracker-banner-count {
	color: var(--grey-500);
	white-space: nowrap;
}

.subproject-tracker-banner-link {
	display: inline-flex;
	align-items: center;
	gap: .25rem;
	flex: 0 0 auto;
	color: var(--link);
	white-space: nowrap;

	&:hover {
		color: var(--link-hover);
	}
}
</style>
