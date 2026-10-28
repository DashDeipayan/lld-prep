// Learner state: loaded from /api/state, mutated through update(), and saved
// back (debounced) to data/progress.json by the server.
import { api } from './api.js';

const defaults = () => ({
  version: 1,
  lang: 'java',
  startDate: null,
  done: {},          // "type:id" -> ISO timestamp
  quiz: {},          // id -> { best, last, total, at }
  bookmarks: [],     // { key, type, id, title, anchor, label, at }
  tutorialStep: {},  // "id:lang" -> highest passed step
  passed: {},        // "kind:id:lang" -> ISO timestamp (all tests green)
  mocks: [],         // attempts
});

let state = defaults();
const listeners = new Set();
let saveTimer = null;

export async function loadState() {
  try {
    state = { ...defaults(), ...(await api.state()) };
  } catch (e) {
    console.error(e);
  }
  if (!state.startDate) {
    state.startDate = new Date().toISOString().slice(0, 10);
    persist();
  }
  return state;
}

export const getState = () => state;

export function subscribe(fn) {
  listeners.add(fn);
  return () => listeners.delete(fn);
}

function persist() {
  clearTimeout(saveTimer);
  saveTimer = setTimeout(() => api.saveState(state).catch(console.error), 300);
}

/** update(fn) where fn mutates a draft copy of the state. */
export function update(fn) {
  const next = structuredClone(state);
  fn(next);
  state = next;
  persist();
  listeners.forEach((l) => l(state));
}

export function replaceState(s) {
  state = { ...defaults(), ...s };
  persist();
  listeners.forEach((l) => l(state));
}

export const resetState = () => replaceState({ startDate: new Date().toISOString().slice(0, 10) });

// ---- helpers ----

export const itemKey = (type, id) => `${type}:${id}`;
export const isDone = (type, id) => !!state.done[itemKey(type, id)];

export function markDone(type, id, done = true) {
  update((s) => {
    const k = itemKey(type, id);
    if (done) s.done[k] = s.done[k] || new Date().toISOString();
    else delete s.done[k];
  });
}

export function isBookmarked(key) {
  return state.bookmarks.some((b) => b.key === key);
}

export function toggleBookmark(b) {
  update((s) => {
    const i = s.bookmarks.findIndex((x) => x.key === b.key);
    if (i >= 0) s.bookmarks.splice(i, 1);
    else s.bookmarks.unshift({ ...b, at: new Date().toISOString() });
  });
}

/** Day number (1-based) the learner is on according to the calendar. */
export function calendarDay() {
  if (!state.startDate) return 1;
  const start = new Date(state.startDate + 'T00:00:00');
  const diff = Math.floor((Date.now() - start.getTime()) / 86400000);
  return Math.max(1, diff + 1);
}
