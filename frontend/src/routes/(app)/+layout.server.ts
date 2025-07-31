import { validateSessionToken, sessionCookieName } from '$lib/server/auth';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies }) => {
    const token = cookies.get(sessionCookieName);
    if (!token) return { user: null };

    const { user, session } = await validateSessionToken(token);
    return { user }; // exposed to layout & children
};