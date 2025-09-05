import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = (async ({ url }) => {
	//const isValid = await validateSession(sessionCookie);
	if (true) {
		// Store the current URL as a redirect parameter
		const redirectUrl = encodeURIComponent(url.pathname + url.search);
		throw redirect(303, `/login?redirect=${redirectUrl}`);
	}

	return {
		user: {
			// Return user data if needed
			authenticated: true
		}
	};
}) satisfies PageLoad;
