<template>
	<svg
		class="progress-ring"
		:class="{'is-all-done': percent >= 1}"
		width="14"
		height="14"
		role="img"
		:aria-label="`${Math.round(percent * 100)}%`"
	>
		<circle
			stroke-width="2"
			fill="transparent"
			cx="50%"
			cy="50%"
			r="5"
		/>
		<circle
			stroke-width="2"
			stroke-dasharray="31.4"
			:stroke-dashoffset="dashOffset"
			stroke-linecap="round"
			fill="transparent"
			cx="50%"
			cy="50%"
			r="5"
		/>
	</svg>
</template>

<script setup lang="ts">
import {computed} from 'vue'

const props = defineProps<{
	percent: number
}>()

const dashOffset = computed(() => {
	const r = 5
	const c = Math.PI * (r * 2)
	return (1 - Math.min(props.percent, 1)) * c
})
</script>

<style scoped lang="scss">
.progress-ring {
	flex-shrink: 0;
	transform: rotate(-90deg);
}

circle {
	stroke: var(--grey-400);

	&:last-child {
		stroke: var(--primary);
		transition: stroke-dashoffset 0.35s;
	}
}

.progress-ring.is-all-done circle:last-child {
	stroke: var(--success);
}
</style>
