import axios from 'axios';
import {
  writable
} from 'svelte/store';

export const loggedIn = writable(false);
export const accountId = writable(undefined);
export const nickname = writable(undefined);
export const token = writable(undefined);
export const dataUrl = writable(undefined);
export const shipInfo = writable(undefined);
export const realm = writable(undefined);
export const statistics = writable([{
    Type: 1,
    Amount: 0,
    Earned: 0,
  },
  {
    Type: 2,
    Amount: 0,
    Earned: 0,
  },
  {
    Type: 3,
    Amount: 0,
    Earned: 0,
  },
]);

export const resourceName = [
  "Republic Tokens",
  "Coal",
  "Steel",
  "Santa Container",
  "Super Container",
  "Anniversary Camouflage",
  "Anniversary Container",
];

axios.get('/warships.min.json').then(res => {
  // transform from array to map
  const ships = {};

  res.data.forEach(s => ships[s.ship_id] = s);
  shipInfo.set(ships);
});

axios.get(`/statistics.json?${new Date().toISOString()}`).then(res => {
  if (typeof res.data !== 'object') {
    return;
  }
  statistics.set(res.data);
});