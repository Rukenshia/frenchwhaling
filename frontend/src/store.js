import axios from 'axios';
import {writable} from 'svelte/store';

export const loggedIn = writable(false);
export const accountId = writable(undefined);
export const nickname = writable(undefined);
export const token = writable(undefined);
export const dataUrl = writable(undefined);
export const shipInfo = writable(undefined);
export const realm = writable(undefined);

axios.get('/warships.min.json').then(res => {
  // transform from array to map
  const ships = {};

  res.data.forEach(s => ships[s.ship_id] = s);
  shipInfo.set(ships);
});