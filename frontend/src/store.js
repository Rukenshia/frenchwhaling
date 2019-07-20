import { writable } from 'svelte/store';

export const loggedIn = writable(false);
export const accountId = writable(undefined);
export const nickname = writable(undefined);
export const token = writable(undefined);