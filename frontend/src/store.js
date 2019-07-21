import { writable } from 'svelte/store';
import axios from 'axios';

export const loggedIn = writable(false);
export const accountId = writable(undefined);
export const nickname = writable(undefined);
export const token = writable(undefined);
export const dataUrl = writable(undefined);
export const shipInfo = writable(undefined);

axios.get('/warships.min.json').then(res => {
    shipInfo.set(res.data);
});