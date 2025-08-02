import { hash, verify } from '@node-rs/argon2';
import { encodeBase32LowerCase } from '@oslojs/encoding';
import { type Actions, fail, redirect } from '@sveltejs/kit';
import { eq, and, desc } from 'drizzle-orm';
import * as auth from '$lib/server/auth';
import { db } from '$lib/server/db';
import * as table from '$lib/server/db/schema';
import type { PageServerLoad } from '../$types';

export const load: PageServerLoad = async (event) => {
    if (!event.locals.user) {
        return { history: [] };
    }

    const history = await db
        .select()
        .from(table.imageHistory)
        .where(eq(table.imageHistory.userId, event.locals.user.id))
        .orderBy(desc(table.imageHistory.createdAt));

    return { history };
};

export const actions: Actions = {
    deleteImage: async (event) => {
        const formData = await event.request.formData();
        const imageId = formData.get('imageId');

        // Check if imageId exists and is a valid number
        if (!imageId || isNaN(Number(imageId))) {
            return fail(400, { message: 'Invalid image ID' });
        }

        if (!event.locals.user) {
            return fail(401, { message: 'Unauthorized' });
        }

        await db
            .delete(table.imageHistory)
            .where(
                and(
                    eq(table.imageHistory.id, Number(imageId)),
                    eq(table.imageHistory.userId, event.locals.user.id)
                )
            );

        return { success: true };
    },
};
