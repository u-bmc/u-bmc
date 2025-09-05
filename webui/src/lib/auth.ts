// SPDX-License-Identifier: BSD-3-Clause

/**
 * Validate the current session by asking the backend.
 * The backend should check the HttpOnly cookie and return 200 if valid,
 * or 401/403 if not.
 */
export async function validateSession(): Promise<boolean> {
	try {
		const res = await fetch('/api/auth/validate', {
			method: 'GET',
			credentials: 'include' // <-- send cookies with the request
		});

		if (res.ok) {
			// Optionally parse user info from JSON if your API provides it
			// const data = await res.json();
			return true;
		}

		// 401/403 = not authorized
		return false;
	} catch (error) {
		// Network/other error: treat as invalid session
		return false;
	}
}
