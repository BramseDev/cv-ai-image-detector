import { integer, numeric, pgTable, serial, text, timestamp } from 'drizzle-orm/pg-core';
import { relations } from 'drizzle-orm';

export const user = pgTable('user', {
    id: text('id').primaryKey(),
    age: integer('age'),
    username: text('username').notNull().unique(),
    passwordHash: text('password_hash').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true }).defaultNow().notNull(),
    updatedAt: timestamp('updated_at', { withTimezone: true }).defaultNow().notNull(),
});

export const userRelations = relations(user, ({ many }) => ({
    sessions: many(session),
    history: many(imageHistory),
}));

export const session = pgTable('session', {
    id: text('id').primaryKey(),
    userId: text('user_id').notNull().references(() => user.id),
    expiresAt: timestamp('expires_at', { withTimezone: true, mode: 'date' }).notNull()
});

export const sessionRelations = relations(session, ({ one }) => ({
    user: one(user, {
        fields: [session.userId],
        references: [user.id]
    })
}));

export const imageHistory = pgTable('image_history', {
    id: serial('id').primaryKey(),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    imageUrl: text('image_url').notNull(),
    imageData: text('image_data').notNull(),
    resultScore: numeric('result_score').notNull(),
    resultConfidence: numeric('result_confidence').notNull(),
    createdAt: timestamp('created_at', { withTimezone: true })
        .defaultNow()
        .notNull(),
});

export const imageHistoryRelations = relations(imageHistory, ({ one }) => ({
    user: one(user, {
        fields: [imageHistory.userId],
        references: [user.id],
    }),
}));


export type Session = typeof session.$inferSelect;
export type User = typeof user.$inferSelect;
export type ImageHistory = typeof imageHistory.$inferSelect;