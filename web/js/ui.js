// Shared UI pieces: html tag, router hooks, icons, small components.
import { h } from 'preact';
import { useEffect, useRef, useState } from 'preact/hooks';
import htm from 'htm';
import { getState, subscribe, update, isBookmarked, toggleBookmark } from './state.js';
import { renderMarkdown, enhance } from './markdown.js';
import { LANGS, langLabel } from './langs.js';

export const html = htm.bind(h);

// ---- routing ----
export function parseHash() {
  const raw = location.hash.replace(/^#\/?/, '');
  const [path, query = ''] = raw.split('?');
  return { parts: path.split('/').filter(Boolean).map(decodeURIComponent), query: new URLSearchParams(query) };
}

export function useRoute() {
  const [r, setR] = useState(parseHash());
  useEffect(() => {
    const on = () => setR(parseHash());
    addEventListener('hashchange', on);
    return () => removeEventListener('hashchange', on);
  }, []);
  return r;
}

export const go = (path) => { location.hash = '#/' + path.replace(/^\//, ''); };

export function useAppState() {
  const [s, set] = useState(getState());
  useEffect(() => subscribe(set), []);
  return s;
}

export function setLang(lang) {
  update((s) => { s.lang = lang; });
  document.documentElement.dataset.lang = lang;
}

export function useAsync(fn, deps) {
  const [v, setV] = useState({ loading: true });
  useEffect(() => {
    let live = true;
    setV({ loading: true });
    fn().then((data) => live && setV({ data }), (error) => live && setV({ error }));
    return () => { live = false; };
  }, deps);
  return v;
}

// ---- icons (inline SVG, stroke = currentColor) ----
const P = {
  home: 'M3 10.5 12 3l9 7.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z',
  book: 'M4 4.5A2.5 2.5 0 0 1 6.5 2H20v17H6.5A2.5 2.5 0 0 0 4 21.5zM4 21.5A2.5 2.5 0 0 1 6.5 19H20v3H6.5',
  code: 'm8 6-6 6 6 6M16 6l6 6-6 6',
  layers: 'm12 2 10 5-10 5L2 7zM2 17l10 5 10-5M2 12l10 5 10-5',
  quiz: 'M9.1 9a3 3 0 0 1 5.8 1c0 2-3 3-3 3M12 17h.01M12 22a10 10 0 1 0 0-20 10 10 0 0 0 0 20z',
  timer: 'M12 22a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM12 10v4l2 2M9 2h6',
  bookmark: 'M19 21l-7-5-7 5V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2z',
  search: 'M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.3-4.3',
  settings: 'M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z',
  check: 'M20 6 9 17l-5-5',
  play: 'M6 4l14 8-14 8z',
  flask: 'M9 3h6M10 3v6L4 19a2 2 0 0 0 1.7 3h12.6a2 2 0 0 0 1.7-3L14 9V3M7 15h10',
  sun: 'M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10zM12 1v2M12 21v2M4.2 4.2l1.4 1.4M18.4 18.4l1.4 1.4M1 12h2M21 12h2M4.2 19.8l1.4-1.4M18.4 5.6l1.4-1.4',
  moon: 'M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z',
  chevron: 'm9 18 6-6-6-6',
  left: 'm15 18-6-6 6-6',
  lock: 'M5 11h14v10H5zM8 11V7a4 4 0 0 1 8 0v4',
  bulb: 'M9 18h6M10 22h4M12 2a7 7 0 0 0-4 12.7V16h8v-1.3A7 7 0 0 0 12 2z',
  eye: 'M1 12s4-8 11-8 11 8 11 8-4 8-11 8S1 12 1 12zM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z',
  reset: 'M3 12a9 9 0 1 0 3-6.7L3 8M3 3v5h5',
  x: 'M18 6 6 18M6 6l12 12',
  menu: 'M3 6h18M3 12h18M3 18h18',
};
export const Icon = ({ name, size = 18, fill = false }) => html`
  <svg class="icon" width=${size} height=${size} viewBox="0 0 24 24" fill=${fill ? 'currentColor' : 'none'}
       stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
    <path d=${P[name]} />
  </svg>`;

export const TYPE_META = {
  lesson: { icon: 'book', label: 'Lesson' },
  exercise: { icon: 'code', label: 'Exercise' },
  tutorial: { icon: 'layers', label: 'Build-along' },
  quiz: { icon: 'quiz', label: 'Quiz' },
  mock: { icon: 'timer', label: 'Mock interview' },
};

export const itemHref = (type, id) => `#/${type}/${encodeURIComponent(id)}`;

// ---- components ----

/** Markdown rendered into a div, then enhanced (code highlight, diagrams). */
export function Markdown({ src, class: cls = '' }) {
  const ref = useRef(null);
  const htmlStr = renderMarkdown(src || '');
  useEffect(() => { enhance(ref.current); }, [htmlStr]);
  return html`<div class=${'prose ' + cls} ref=${ref} dangerouslySetInnerHTML=${{ __html: htmlStr }} />`;
}

export function BookmarkButton({ bkey, type, id, title, anchor, label, compact }) {
  useAppState();
  const on = isBookmarked(bkey);
  return html`<button type="button" class=${'btn ghost bookmark-btn' + (on ? ' on' : '')} title=${on ? 'Remove bookmark' : 'Bookmark'}
    aria-pressed=${on}
    onClick=${() => toggleBookmark({ key: bkey, type, id, title, anchor, label })}>
    <${Icon} name="bookmark" fill=${on} />${compact ? null : html`<span>${on ? 'Bookmarked' : 'Bookmark'}</span>`}
  </button>`;
}

export function LangToggle({ lang, onChange }) {
  return html`<div class="seg" role="group" aria-label="Language">
    ${LANGS.map((l) => html`<button type="button" class=${lang === l ? 'on' : ''} aria-pressed=${lang === l}
      onClick=${() => onChange(l)}>${langLabel(l)}</button>`)}
  </div>`;
}

export function Progress({ value, total, size = 'md' }) {
  const pct = total ? Math.round((value / total) * 100) : 0;
  return html`<div class=${'progress ' + size} role="progressbar" aria-valuenow=${pct} aria-valuemin="0" aria-valuemax="100">
    <div class="bar" style=${{ width: pct + '%' }} />
  </div>`;
}

export function Ring({ value, total, size = 34 }) {
  const pct = total ? value / total : 0;
  const r = (size - 5) / 2;
  const c = 2 * Math.PI * r;
  return html`<svg class="ring" width=${size} height=${size} viewBox=${`0 0 ${size} ${size}`}>
    <circle cx=${size / 2} cy=${size / 2} r=${r} class="ring-bg" />
    <circle cx=${size / 2} cy=${size / 2} r=${r} class="ring-fg" stroke-dasharray=${c} stroke-dashoffset=${c * (1 - pct)}
      transform=${`rotate(-90 ${size / 2} ${size / 2})`} />
  </svg>`;
}

export const Loading = () => html`<div class="loading"><div class="spinner" /> Loading…</div>`;
export const ErrorBox = ({ error }) => html`<div class="callout danger">${String(error?.message || error)}</div>`;

export function Difficulty({ level }) {
  if (!level) return null;
  return html`<span class=${'chip diff-' + level}>${level}</span>`;
}

/** Section bookmarking: clicking the marker next to an h2/h3 bookmarks that section. */
export function useSectionBookmarks(ref, { type, id, title }, deps) {
  useEffect(() => {
    const root = ref.current;
    if (!root) return;
    const heads = root.querySelectorAll('h2.anchored, h3.anchored');
    heads.forEach((hd) => {
      if (hd.querySelector('.sec-mark')) return;
      const b = document.createElement('button');
      b.type = 'button';
      b.className = 'sec-mark';
      b.title = 'Bookmark this section';
      const key = `${type}:${id}#${hd.id}`;
      const paint = () => {
        const on = isBookmarked(key);
        b.classList.toggle('on', on);
        b.setAttribute('aria-pressed', on);
        b.innerHTML = `<svg width="15" height="15" viewBox="0 0 24 24" fill="${on ? 'currentColor' : 'none'}" stroke="currentColor" stroke-width="2"><path d="${P.bookmark}"/></svg>`;
      };
      paint();
      b.onclick = () => { toggleBookmark({ key, type, id, title, anchor: hd.id, label: hd.textContent }); };
      hd.appendChild(b);
      hd._paint = paint;
    });
    const unsub = subscribe(() => heads.forEach((hd) => hd._paint && hd._paint()));
    return unsub;
  }, deps);
}

export function scrollToAnchor(anchor) {
  if (!anchor) return;
  setTimeout(() => document.getElementById(anchor)?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 150);
}

export const fmtMin = (m) => (m >= 60 ? `${Math.floor(m / 60)}h${m % 60 ? ' ' + (m % 60) + 'm' : ''}` : `${m}m`);
