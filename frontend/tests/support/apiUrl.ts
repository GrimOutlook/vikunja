/**
 * The api base url without a trailing slash. `mage test:e2e` sets API_URL with
 * one, so helpers appending a path would otherwise build urls like
 * `/api/v1//user`, which the api answers with a 404.
 *
 * Note that the playwright request context in fixtures.ts needs the trailing
 * slash instead: it resolves relative paths against its baseURL.
 */
export function getApiUrl(): string {
	return (process.env.API_URL || 'http://127.0.0.1:3456/api/v1').replace(/\/+$/, '')
}
