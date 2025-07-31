import * as auth from '$lib/server/auth';
import { db } from '$lib/server/db';
import * as table from '$lib/server/db/schema';
import { type Actions, fail, redirect } from '@sveltejs/kit';

// export const load: PageServerLoad = async (event) => {
//     if (!event.locals.user) {
//         return redirect(302, '/login');
//     }
//     return {user: event.locals.user};
// };

export const actions: Actions = {
    logout: async (event) => {
        if (!event.locals.session) {
            return fail(401);
        }
        await auth.invalidateSession(event.locals.session.id);
        auth.deleteSessionTokenCookie(event);

        return redirect(302, '/login');
    },
    saveImage: async (event) => {
        if (!event.locals.user) {
            return fail(401, { message: 'Unauthorized' });
        }

        const formData = await event.request.formData();
        const file = formData.get('image');
        const resultScore = Number(formData.get('result_score'));
        const resultConfidence = Number(formData.get('result_confidence'));

        if (!(file instanceof File)) {
            return fail(400, { message: 'No file uploaded' });
        }

        const buffer = await file.arrayBuffer();
        const base64Image = bufferToBase64(buffer);

        const filename = `${file.name}`;

        // Optional: store image history in DB
        await db.insert(table.imageHistory).values({
            userId: event.locals.user.id,
            imageUrl: `/uploads/${filename}`,
            resultScore: resultScore,
            resultConfidence: resultConfidence,
            imageData: base64Image // Save the Base64 encoded image
        });

        return { success: true };
    },
};

// Helper function to convert a buffer to a Base64 string
function bufferToBase64(buffer: ArrayBuffer): string {
    const uint8Array = new Uint8Array(buffer);
    let binary = '';
    uint8Array.forEach((byte) => {
        binary += String.fromCharCode(byte);
    });
    return `data:image/png;base64,${btoa(binary)}`; // assuming PNG image
}