import {describe, it, expect, vi, beforeEach} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import en from '@/i18n/lang/en.json'

const countAll = vi.fn()

vi.mock('@/services/task', () => ({
	default: class {
		countAll = countAll
	},
}))

const TaskCountStats = (await import('./TaskCountStats.vue')).default

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

async function mountStats() {
	const wrapper = mount(TaskCountStats, {global: {plugins: [i18n]}})
	await flushPromises()
	return wrapper
}

describe('TaskCountStats', () => {
	beforeEach(() => countAll.mockClear())

	it('shows the number of open and completed tasks', async () => {
		countAll.mockImplementation((filter: string) => Promise.resolve(filter === 'done = false' ? 42 : 7))

		const wrapper = await mountStats()

		expect(countAll).toHaveBeenCalledWith('done = false')
		expect(countAll).toHaveBeenCalledWith('done = true')
		const stats = wrapper.findAll('.task-count-stats__stat')
		expect(stats).toHaveLength(2)
		expect(stats[0].text()).toContain('42')
		expect(stats[0].text()).toContain('Open tasks')
		expect(stats[1].text()).toContain('7')
		expect(stats[1].text()).toContain('Completed tasks')
	})

	it('renders nothing when the counts cannot be loaded', async () => {
		countAll.mockImplementation((filter: string) => filter === 'done = false'
			? Promise.reject(new Error('nope'))
			: Promise.resolve(7))

		const wrapper = await mountStats()

		expect(wrapper.find('.task-count-stats').exists()).toBe(false)
	})
})
