import {describe, it, expect, beforeEach, vi} from 'vitest'
import {defineComponent, h, nextTick} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createRouter, createMemoryHistory, type Router} from 'vue-router'
import {createI18n} from 'vue-i18n'
import en from '@/i18n/lang/en.json'

const getAll = vi.fn(async () => [])
vi.mock('@/services/taskCollection', async (importOriginal) => {
	const actual = await importOriginal<typeof import('@/services/taskCollection')>()
	return {
		...actual,
		default: class {
			loading = false
			totalPages = 1
			getAll = getAll
		},
	}
})

import {useTaskList, buildStoredQuery} from './useTaskList'

describe('buildStoredQuery', () => {
	it('includes sort when set', () => {
		expect(buildStoredQuery({sort: 'due_date:asc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'due_date:asc'})
	})

	it('includes filter and search when set', () => {
		expect(buildStoredQuery({sort: undefined, filter: 'done = false', s: 'foo', page: 1}))
			.toEqual({filter: 'done = false', s: 'foo'})
	})

	it('omits page when it equals the default of 1', () => {
		expect(buildStoredQuery({sort: 'id:desc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'id:desc'})
	})

	it('includes page when greater than 1', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 3}))
			.toEqual({page: '3'})
	})

	it('returns an empty object when nothing is set', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 1}))
			.toEqual({})
	})

	it('skips empty strings', () => {
		expect(buildStoredQuery({sort: '', filter: '', s: '', page: 1}))
			.toEqual({})
	})
})

// The second positional argument passed to TaskCollectionService.getAll carries
// the sort_by/order_by the backend uses to decide whether to rank by relevance.
function lastRequestParams(): Record<string, unknown> {
	return getAll.mock.calls.at(-1)?.[1] as Record<string, unknown>
}

async function mountTaskList(query: Record<string, string>): Promise<Router> {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/', name: 'home', component: {render: () => null}}],
	})
	await router.push({path: '/', query})
	await router.isReady()

	const TestComponent = defineComponent({
		setup() {
			useTaskList(() => 1, () => 1)
			return () => h('div')
		},
	})

	// The base store calls useI18n() when pinia instantiates it, which needs a
	// global i18n instance to be installed on the app.
	const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

	mount(TestComponent, {global: {plugins: [router, i18n]}})
	await flushPromises()
	await nextTick()
	return router
}

describe('useTaskList sort handling for relevance ranking', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		getAll.mockClear()
	})

	it('omits the sort while searching with the default sort so the backend ranks by relevance', async () => {
		await mountTaskList({s: 'find me'})

		const params = lastRequestParams()
		expect(params.s).toBe('find me')
		expect(params.sort_by).toEqual([])
		expect(params.order_by).toEqual([])
	})

	it('keeps an explicit user sort while searching so the user sort is respected', async () => {
		await mountTaskList({s: 'find me', sort: 'title:asc'})

		const params = lastRequestParams()
		expect(params.s).toBe('find me')
		expect(params.sort_by).toEqual(['title'])
		expect(params.order_by).toEqual(['asc'])
	})

	it('sends the default sort when not searching', async () => {
		await mountTaskList({})

		const params = lastRequestParams()
		expect(params.s).toBe('')
		expect(params.sort_by).not.toHaveLength(0)
		// id always sorts last so other sort columns take precedence.
		expect(params.sort_by).toEqual(['id'])
		expect(params.order_by).toEqual(['desc'])
	})
})
