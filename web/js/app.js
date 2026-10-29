import { render } from 'preact';
import { useEffect, useRef, useState } from 'preact/hooks';
import { api } from './api.js';
import { loadState, getState, update, markDone, isDone, calendarDay, replaceState, resetState, itemKey } from './state.js';
import {
  html, useRoute, useAppState, useAsync, go, setLang, Icon, TYPE_META, itemHref, Markdown, BookmarkButton,
  LangToggle, Progress, Ring, Loading, ErrorBox, Difficulty, useSectionBookmarks, scrollToAnchor, fmtMin,
} from './ui.js';
import { applyTheme } from './editor.js';
import { resetMermaidTheme } from './markdown.js';
import { ExerciseView, TutorialView } from './code.js';
import { QuizView } from './quiz.js';
import { MocksView, MockView } from './mock.js';
import { SignIn, ProviderButton, PROVIDER_LABEL, startProvider, takeHashParam } from './auth.js';

let PLAN = null;
export const plan = () => PLAN;
// { mode: 'local' } or { mode: 'accounts', user, inviteRequired } from /api/auth/me.
let AUTH = { mode: 'local' };

// ---- plan helpers ----
function flatItems() {
  return PLAN.days.flatMap((d) => d.items.map((it) => ({ ...it, day: d.day })));
}
export function locate(type, id) {
  const all = flatItems();
  const i = all.findIndex((x) => x.type === type && x.id === id);
  return i < 0 ? null : { item: all[i], next: all[i + 1] || null, prev: all[i - 1] || null };
}
const dayStats = (d) => ({ done: d.items.filter((it) => isDone(it.type, it.id)).length, total: d.items.length });
function currentDay() {
  // First day with unfinished work, but never ahead of where the learner has got to.
  const firstOpen = PLAN.days.find((d) => dayStats(d).done < d.items.length);
  return firstOpen ? firstOpen.day : PLAN.days.length;
}

// ---- shell ----
function Sidebar({ route, open, onClose }) {
  useAppState();
  const top = route.parts[0] || '';
  const cur = currentDay();
  const link = (href, icon, label, active) => html`
    <a href=${href} class=${'nav-link' + (active ? ' active' : '')} onClick=${onClose}><${Icon} name=${icon} /><span>${label}</span></a>`;
  return html`<aside class=${'sidebar' + (open ? ' open' : '')}>
    <a class="brand" href="#/" onClick=${onClose}>
      <span class="logo"><svg viewBox="0 0 32 32" width="26" height="26"><rect width="32" height="32" rx="7" fill="currentColor"/><path d="M9 9h5v5H9zM18 9h5v5h-5zM13.5 18h5v5h-5z" fill="var(--bg)"/></svg></span>
      <span><b>LLD Prep</b><small>14 days to the offer</small></span>
    </a>
    <nav>
      ${link('#/', 'home', 'Today', top === '')}
      ${link('#/library', 'book', 'Library', top === 'library')}
      ${link('#/mocks', 'timer', 'Mock interviews', top === 'mocks' || top === 'mock')}
      ${link('#/bookmarks', 'bookmark', 'Bookmarks', top === 'bookmarks')}
      ${link('#/settings', 'settings', 'Settings & data', top === 'settings')}
    </nav>
    <div class="nav-heading">The plan</div>
    <div class="days">
      ${PLAN.days.map((d) => {
        const st = dayStats(d);
        const active = top === 'day' && Number(route.parts[1]) === d.day;
        return html`<a href=${'#/day/' + d.day} onClick=${onClose}
          class=${'day-link' + (active ? ' active' : '') + (st.done === st.total ? ' complete' : '') + (d.day === cur ? ' current' : '')}>
          <span class="day-num">${st.done === st.total ? html`<${Icon} name="check" size=${13} />` : d.day}</span>
          <span class="day-title">${d.title}</span>
        </a>`;
      })}
    </div>
  </aside>`;
}

function Topbar({ onMenu }) {
  const s = useAppState();
  const [q, setQ] = useState('');
  const input = useRef(null);
  useEffect(() => {
    const onKey = (e) => {
      if ((e.key === 'k' && (e.metaKey || e.ctrlKey)) || (e.key === '/' && !/INPUT|TEXTAREA/.test(document.activeElement?.tagName) && !document.activeElement?.closest('.monaco-editor'))) {
        e.preventDefault();
        input.current?.focus();
      }
    };
    addEventListener('keydown', onKey);
    return () => removeEventListener('keydown', onKey);
  }, []);
  const [theme, setTheme] = useState(document.documentElement.dataset.theme || (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'));
  const flipTheme = () => {
    const t = theme === 'dark' ? 'light' : 'dark';
    document.documentElement.dataset.theme = t;
    try { localStorage.setItem('lldprep.theme', JSON.stringify(t)); } catch (e) {}
    setTheme(t);
    applyTheme();
    resetMermaidTheme();
  };
  return html`<header class="topbar">
    <button class="btn ghost icon-only menu-btn" onClick=${onMenu} aria-label="Menu"><${Icon} name="menu" /></button>
    <form class="search" role="search" onSubmit=${(e) => { e.preventDefault(); if (q.trim()) go('search?q=' + encodeURIComponent(q.trim())); }}>
      <${Icon} name="search" size=${16} />
      <input ref=${input} value=${q} onInput=${(e) => setQ(e.target.value)} placeholder="Search patterns, problems, concepts…" aria-label="Search" />
      <kbd>⌘K</kbd>
    </form>
    <div class="topbar-right">
      <${LangToggle} lang=${s.lang} onChange=${setLang} />
      <button class="btn ghost icon-only" onClick=${flipTheme} title="Toggle theme" aria-label="Toggle theme">
        <${Icon} name=${theme === 'dark' ? 'sun' : 'moon'} />
      </button>
    </div>
  </header>`;
}

// ---- pages ----
function ItemRow({ it, showDay }) {
  useAppState();
  const done = isDone(it.type, it.id);
  const meta = TYPE_META[it.type];
  return html`<div class=${'item-row' + (done ? ' done' : '')}>
    <button class="check" title=${done ? 'Mark as not done' : 'Mark as done'} aria-pressed=${done}
      onClick=${() => markDone(it.type, it.id, !done)}>${done ? html`<${Icon} name="check" size=${14} />` : null}</button>
    <a class="item-main" href=${itemHref(it.type, it.id)}>
      <span class=${'type-badge t-' + it.type}><${Icon} name=${meta.icon} size=${14} />${meta.label}</span>
      <span class="item-title">${it.title}</span>
      ${it.note ? html`<span class="item-note">${it.note}</span>` : null}
    </a>
    ${showDay ? html`<span class="muted small">Day ${it.day}</span>` : null}
    <span class="item-min">${fmtMin(it.minutes)}</span>
  </div>`;
}

function Dashboard() {
  const s = useAppState();
  const all = flatItems();
  const doneCount = all.filter((it) => isDone(it.type, it.id)).length;
  const cur = currentDay();
  const day = PLAN.days[cur - 1];
  const st = dayStats(day);
  const next = all.find((it) => !isDone(it.type, it.id));
  const cal = calendarDay();
  const passedEx = Object.keys(s.passed).length;
  const quizzes = Object.values(s.quiz);
  const quizAvg = quizzes.length ? Math.round(quizzes.reduce((a, q) => a + q.best / q.total, 0) / quizzes.length * 100) : null;
  const mocks = s.mocks.filter((m) => m.score != null);
  const remaining = day.items.filter((it) => !isDone(it.type, it.id)).reduce((a, it) => a + it.minutes, 0);
  const behind = cal - cur;

  return html`<div class="page dashboard">
    <section class="hero">
      <div>
        <div class="eyebrow">Day ${cur} of ${PLAN.days.length}${behind > 0 ? html` · <span class="warn-text">${behind} day${behind > 1 ? 's' : ''} behind calendar</span>` : null}</div>
        <h1>${day.title}</h1>
        <p class="lead">${day.goal}</p>
        <div class="hero-actions">
          ${next ? html`<a class="btn primary lg" href=${itemHref(next.type, next.id)}><${Icon} name="play" size=${16} />
            ${doneCount === 0 ? 'Start' : 'Continue'}: ${next.title}</a>` : html`<a class="btn primary lg" href="#/mocks">All done. Do another mock</a>`}
          <a class="btn" href=${'#/day/' + cur}>Day ${cur} agenda</a>
        </div>
      </div>
      <div class="hero-stat">
        <${Ring} value=${doneCount} total=${all.length} size=${120} />
        <div class="ring-label"><b>${Math.round((doneCount / all.length) * 100)}%</b><span>${doneCount}/${all.length} items</span></div>
      </div>
    </section>

    <section class="stats">
      <div class="stat"><span class="stat-v">${st.done}/${st.total}</span><span class="stat-l">today's items</span></div>
      <div class="stat"><span class="stat-v">${fmtMin(remaining)}</span><span class="stat-l">left today</span></div>
      <div class="stat"><span class="stat-v">${passedEx}</span><span class="stat-l">solutions passing</span></div>
      <div class="stat"><span class="stat-v">${quizAvg == null ? '–' : quizAvg + '%'}</span><span class="stat-l">quiz average</span></div>
      <div class="stat"><span class="stat-v">${mocks.length}</span><span class="stat-l">mocks scored</span></div>
    </section>

    <section class="card">
      <div class="card-head"><h2>Today · Day ${cur}</h2><${Progress} value=${st.done} total=${st.total} /></div>
      <div class="items">${day.items.map((it) => html`<${ItemRow} it=${{ ...it, day: cur }} />`)}</div>
    </section>

    <section>
      <h2 class="section-title">The 14-day plan</h2>
      <div class="plan-grid">
        ${PLAN.days.map((d) => {
          const x = dayStats(d);
          return html`<a class=${'plan-cell' + (x.done === x.total ? ' complete' : '') + (d.day === cur ? ' current' : '')} href=${'#/day/' + d.day}>
            <div class="plan-cell-top"><span class="day-pill">Day ${d.day}</span><span class="muted small">${x.done}/${x.total}</span></div>
            <div class="plan-cell-title">${d.title}</div>
            <${Progress} value=${x.done} total=${x.total} size="sm" />
          </a>`;
        })}
      </div>
    </section>
  </div>`;
}

function DayView({ n }) {
  useAppState();
  const d = PLAN.days[n - 1];
  if (!d) return html`<div class="page"><${ErrorBox} error="No such day" /></div>`;
  const st = dayStats(d);
  const total = d.items.reduce((a, it) => a + it.minutes, 0);
  return html`<div class="page narrow">
    <div class="crumbs"><a href="#/">Plan</a> / Day ${d.day}</div>
    <h1>Day ${d.day}: ${d.title}</h1>
    <p class="lead">${d.goal}</p>
    <div class="meta-row"><span>${fmtMin(total)} planned</span><span>${st.done}/${st.total} done</span></div>
    <${Progress} value=${st.done} total=${st.total} />
    <div class="items spaced">${d.items.map((it) => html`<${ItemRow} it=${{ ...it, day: d.day }} />`)}</div>
    <div class="pager">
      ${d.day > 1 ? html`<a class="btn" href=${'#/day/' + (d.day - 1)}><${Icon} name="left" size=${16} /> Day ${d.day - 1}</a>` : html`<span />`}
      ${d.day < PLAN.days.length ? html`<a class="btn" href=${'#/day/' + (d.day + 1)}>Day ${d.day + 1} <${Icon} name="chevron" size=${16} /></a>` : null}
    </div>
  </div>`;
}

export function NextUp({ type, id, onDone }) {
  useAppState();
  const loc = locate(type, id);
  const done = isDone(type, id);
  return html`<div class="next-up">
    ${onDone !== false ? html`<button class=${'btn' + (done ? '' : ' primary')} onClick=${() => markDone(type, id, !done)}>
      <${Icon} name="check" size=${16} /> ${done ? 'Completed' : 'Mark complete'}</button>` : null}
    ${loc?.next ? html`<a class="btn ghost next-link" href=${itemHref(loc.next.type, loc.next.id)}
        onClick=${() => { if (onDone !== false && !done) markDone(type, id, true); }}>
      <span class="muted small">Next${loc.next.day !== loc.item.day ? ` · Day ${loc.next.day}` : ''}</span>
      <span>${loc.next.title}</span><${Icon} name="chevron" size=${16} /></a>` : null}
  </div>`;
}

export function Crumbs({ type, id }) {
  const loc = locate(type, id);
  return html`<div class="crumbs">
    <a href="#/">Plan</a> / ${loc ? html`<a href=${'#/day/' + loc.item.day}>Day ${loc.item.day}</a> / ` : html`<a href="#/library">Library</a> / `}
    <span>${TYPE_META[type].label}</span>
  </div>`;
}

function LessonView({ id, anchor }) {
  const r = useAsync(() => api.lesson(id), [id]);
  const ref = useRef(null);
  useSectionBookmarks(ref, { type: 'lesson', id, title: r.data?.title }, [r.data]);
  useEffect(() => { if (r.data) scrollToAnchor(anchor); else window.scrollTo(0, 0); }, [r.data, anchor]);
  if (r.loading) return html`<${Loading} />`;
  if (r.error) return html`<${ErrorBox} error=${r.error} />`;
  const l = r.data;
  return html`<div class="page narrow lesson" ref=${ref}>
    <${Crumbs} type="lesson" id=${id} />
    <div class="title-row"><h1>${l.title}</h1><${BookmarkButton} bkey=${'lesson:' + id} type="lesson" id=${id} title=${l.title} /></div>
    ${l.summary ? html`<p class="lead">${l.summary}</p>` : null}
    <${Markdown} src=${l.markdown} />
    <${NextUp} type="lesson" id=${id} />
  </div>`;
}

function Library() {
  useAppState();
  const c = PLAN.catalog;
  const [tab, setTab] = useState('exercises');
  const tabs = [['exercises', 'Exercises', 'exercise'], ['tutorials', 'Build-alongs', 'tutorial'], ['lessons', 'Lessons', 'lesson'], ['quizzes', 'Quizzes', 'quiz']];
  const type = tabs.find((t) => t[0] === tab)[2];
  return html`<div class="page">
    <h1>Library</h1>
    <p class="lead">Everything in the course, outside the day-by-day plan.</p>
    <div class="tabs">${tabs.map(([k, label]) => html`<button class=${tab === k ? 'on' : ''} onClick=${() => setTab(k)}>${label} <span class="count">${c[k].length}</span></button>`)}</div>
    <div class="lib-grid">
      ${c[tab].map((e) => {
        const loc = locate(type, e.ID);
        const done = isDone(type, e.ID);
        return html`<a class=${'lib-card' + (done ? ' done' : '')} href=${itemHref(type, e.ID)}>
          <div class="lib-card-top">
            <span class=${'type-badge t-' + type}><${Icon} name=${TYPE_META[type].icon} size=${14} />${TYPE_META[type].label}</span>
            ${done ? html`<span class="chip ok"><${Icon} name="check" size=${12} /> done</span>` : null}
          </div>
          <h3>${e.Title}</h3>
          ${e.Summary ? html`<p>${e.Summary}</p>` : null}
          <div class="lib-meta">
            <${Difficulty} level=${e.Difficulty} />
            ${e.Minutes ? html`<span>${fmtMin(e.Minutes)}</span>` : null}
            ${loc ? html`<span>Day ${loc.item.day}</span>` : null}
            ${(e.Tags || []).slice(0, 3).map((t) => html`<span class="tag">${t}</span>`)}
          </div>
        </a>`;
      })}
    </div>
  </div>`;
}

function Bookmarks() {
  const s = useAppState();
  return html`<div class="page narrow">
    <h1>Bookmarks</h1>
    <p class="lead">Bookmark whole pages with the button in a page header, or single sections with the marker that appears when you hover a heading.</p>
    ${s.bookmarks.length === 0 ? html`<div class="empty">No bookmarks yet.</div>` : html`<div class="items spaced">
      ${s.bookmarks.map((b) => html`<div class="item-row">
        <a class="item-main" href=${itemHref(b.type, b.id) + (b.anchor ? '?a=' + encodeURIComponent(b.anchor) : '')}>
          <span class=${'type-badge t-' + b.type}><${Icon} name=${TYPE_META[b.type]?.icon || 'book'} size=${14} />${TYPE_META[b.type]?.label}</span>
          <span class="item-title">${b.label ? html`${b.label} <span class="muted">in ${b.title}</span>` : b.title}</span>
        </a>
        <span class="muted small">${new Date(b.at).toLocaleDateString()}</span>
        <button class="btn ghost icon-only" title="Remove" onClick=${() => update((st) => { st.bookmarks = st.bookmarks.filter((x) => x.key !== b.key); })}><${Icon} name="x" size=${16} /></button>
      </div>`)}
    </div>`}
  </div>`;
}

function SearchView({ q }) {
  const r = useAsync(() => api.search(q), [q]);
  const mark = (text) => {
    const terms = q.toLowerCase().split(/\s+/).filter(Boolean).map((t) => t.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'));
    if (!terms.length) return text;
    const parts = text.split(new RegExp(`(${terms.join('|')})`, 'gi'));
    return parts.map((p, i) => (i % 2 ? html`<mark>${p}</mark>` : p));
  };
  return html`<div class="page narrow">
    <h1>Search</h1>
    <p class="lead">Results for “${q}”</p>
    ${r.loading ? html`<${Loading} />` : r.error ? html`<${ErrorBox} error=${r.error} />` : r.data.length === 0 ? html`<div class="empty">Nothing matched. Try a pattern name (“observer”), a problem (“elevator”) or a concept (“deadlock”).</div>` : html`
      <div class="search-results">${r.data.map((h) => html`<a class="search-hit" href=${itemHref(h.type, h.id)}>
        <span class=${'type-badge t-' + h.type}><${Icon} name=${TYPE_META[h.type].icon} size=${14} />${TYPE_META[h.type].label}</span>
        <div class="hit-title">${mark(h.title)}</div>
        <div class="hit-snippet">${mark(h.snippet)}</div>
      </a>`)}</div>`}
  </div>`;
}

function Account() {
  const [busy, setBusy] = useState(false);
  // Set by the server after "Connect Google/GitHub" (success or failure).
  const [notice] = useState(() => {
    const connected = takeHashParam('connected');
    const error = takeHashParam('auth_error');
    if (error) return { bad: true, text: error };
    if (connected) return { bad: false, text: `${PROVIDER_LABEL[connected] || connected} is connected. You can now sign in with it.` };
    return null;
  });
  const signOut = async () => {
    setBusy(true);
    await api.logout().catch(() => {});
    location.reload();
  };
  const linked = AUTH.linked || [];
  const providers = AUTH.providers || [];
  return html`<section class="card">
    <h2>Account</h2>
    <div class="account-row">
      <span>Signed in as <b>${AUTH.user}</b></span>
      <button class="btn" onClick=${signOut} disabled=${busy}>Sign out</button>
    </div>
    ${notice ? html`<p class=${notice.bad ? 'auth-error' : 'auth-ok'} role="status">${notice.text}</p>` : null}
    ${providers.length ? html`
      <h3 class="account-sub">Sign-in methods</h3>
      <ul class="sign-in-methods">
        <li><span>Username and password</span>${AUTH.hasPassword ? html`<span class="chip ok">On</span>` : html`<span class="muted small">Not set (you signed up with ${linked.map((p) => PROVIDER_LABEL[p]).join(' / ')})</span>`}</li>
        ${providers.map((p) => html`<li>
          <span>${PROVIDER_LABEL[p]}</span>
          ${linked.includes(p) ? html`<span class="chip ok">Connected</span>`
            : html`<${ProviderButton} provider=${p} label=${'Connect ' + PROVIDER_LABEL[p]} onClick=${() => startProvider(p, { mode: 'link' })} />`}
        </li>`)}
      </ul>` : null}
  </section>`;
}

function Settings() {
  const s = useAppState();
  const tc = useAsync(() => api.toolchains(), []);
  const fileRef = useRef(null);
  const [msg, setMsg] = useState('');
  const exportData = () => {
    const blob = new Blob([JSON.stringify(getState(), null, 2)], { type: 'application/json' });
    const a = document.createElement('a');
    a.href = URL.createObjectURL(blob);
    a.download = `lldprep-progress-${new Date().toISOString().slice(0, 10)}.json`;
    a.click();
    URL.revokeObjectURL(a.href);
  };
  const importData = async (e) => {
    const f = e.target.files[0];
    if (!f) return;
    try {
      replaceState(JSON.parse(await f.text()));
      setMsg('Progress imported.');
    } catch (err) {
      setMsg('Could not import: ' + err.message);
    }
    e.target.value = '';
  };
  return html`<div class="page narrow">
    <h1>Settings & data</h1>
    ${AUTH.mode === 'accounts' ? html`<${Account} />` : null}
    <section class="card">
      <h2>Preferences</h2>
      <div class="form-row"><label>Default language</label><${LangToggle} lang=${s.lang} onChange=${setLang} /></div>
      <div class="form-row"><label for="start">Plan start date</label>
        <input id="start" type="date" value=${s.startDate} onChange=${(e) => update((st) => { st.startDate = e.target.value; })} />
        <span class="muted small">Used to tell you if you're behind the calendar.</span></div>
    </section>
    <section class="card">
      <h2>Toolchains</h2>
      <p class="muted">${AUTH.mode === 'accounts' ? 'Code runs on the server with its installed compilers.' : 'Code runs on your machine with the locally installed compilers.'}</p>
      ${tc.data ? html`<ul class="tc">
        <li class=${tc.data.java ? 'ok' : 'bad'}><b>Java</b> ${tc.data.java || 'not found. Install JDK 17+ and restart.'}</li>
        <li class=${tc.data.go ? 'ok' : 'bad'}><b>Go</b> ${tc.data.go || 'not found. Install Go 1.23+ and restart.'}</li>
        <li class=${tc.data.python ? 'ok' : 'bad'}><b>Python</b> ${tc.data.python || 'not found. Install Python 3.8+ and restart.'}</li>
      </ul>` : html`<${Loading} />`}
    </section>
    <section class="card">
      <h2>Your data</h2>
      ${AUTH.mode === 'accounts'
        ? html`<p class="muted">Your progress and code are saved to your account. Export makes a portable backup, and Import moves progress over from another copy of the app.</p>`
        : html`<p class="muted">Progress lives in <code>data/progress.json</code> and your code in <code>data/drafts/</code>, next to the app. Export makes a portable backup.</p>`}
      <div class="btn-row">
        <button class="btn" onClick=${exportData}>Export progress</button>
        <button class="btn" onClick=${() => fileRef.current.click()}>Import progress</button>
        <input ref=${fileRef} type="file" accept="application/json" hidden onChange=${importData} />
        <button class="btn danger" onClick=${() => { if (confirm('Reset all progress, quiz scores, bookmarks and mock history? Code drafts are kept.')) { resetState(); setMsg('Progress reset.'); } }}>Reset progress</button>
      </div>
      ${msg ? html`<p class="muted">${msg}</p>` : null}
    </section>
    <section class="card">
      <h2>Keyboard</h2>
      <ul class="kbd-list">
        <li><kbd>⌘/Ctrl K</kbd> or <kbd>/</kbd> search</li>
        <li><kbd>⌘/Ctrl '</kbd> run code</li>
        <li><kbd>⌘/Ctrl Enter</kbd> run tests</li>
      </ul>
    </section>
  </div>`;
}

function App() {
  const route = useRoute();
  const [menu, setMenu] = useState(false);
  const [p0, p1, p2] = route.parts;
  const anchor = route.query.get('a');
  useEffect(() => { if (!/^(exercise|tutorial|mock)$/.test(p0)) window.scrollTo(0, 0); }, [p0, p1]);
  let page;
  switch (p0) {
    case undefined: page = html`<${Dashboard} />`; break;
    case 'day': page = html`<${DayView} n=${Number(p1)} />`; break;
    case 'lesson': page = html`<${LessonView} id=${p1} anchor=${anchor} key=${p1} />`; break;
    case 'exercise': page = html`<${ExerciseView} id=${p1} key=${p1} />`; break;
    case 'tutorial': page = html`<${TutorialView} id=${p1} key=${p1} anchor=${anchor} />`; break;
    case 'quiz': page = html`<${QuizView} id=${p1} key=${p1} />`; break;
    case 'mocks': page = html`<${MocksView} />`; break;
    case 'mock': page = html`<${MockView} id=${p1} key=${p1} />`; break;
    case 'library': page = html`<${Library} />`; break;
    case 'bookmarks': page = html`<${Bookmarks} />`; break;
    case 'search': page = html`<${SearchView} q=${route.query.get('q') || ''} key=${route.query.get('q')} />`; break;
    case 'settings': page = html`<${Settings} />`; break;
    default: page = html`<div class="page"><${ErrorBox} error="Page not found" /></div>`;
  }
  const wide = /^(exercise|tutorial|mock)$/.test(p0);
  return html`<div class=${'shell' + (wide ? ' wide' : '')}>
    <${Sidebar} route=${route} open=${menu} onClose=${() => setMenu(false)} />
    ${menu ? html`<div class="scrim" onClick=${() => setMenu(false)} />` : null}
    <div class="main">
      <${Topbar} onMenu=${() => setMenu(true)} />
      <main class="content">${page}</main>
    </div>
  </div>`;
}

// Global: language tab clicks inside rendered markdown switch the language everywhere.
document.addEventListener('click', (e) => {
  const t = e.target.closest('[data-lang-tab]');
  if (t) setLang(t.dataset.langTab);
});

const root = document.getElementById('app');

async function boot() {
  AUTH = await api.me();
  if (AUTH.mode === 'accounts' && !AUTH.user) {
    render(html`<${SignIn} inviteRequired=${AUTH.inviteRequired} providers=${AUTH.providers || []} onSignedIn=${() => boot().catch(bootFailed)} />`, root);
    return;
  }
  const [p, s] = await Promise.all([api.plan(), loadState()]);
  PLAN = p;
  document.documentElement.dataset.lang = s.lang;
  render(html`<${App} />`, root);
}

function bootFailed(e) {
  root.innerHTML = '<div class="boot error"></div>';
  root.firstChild.textContent = `Failed to start: ${e.message}. Is the server running?`;
}

// A 401 from any API call means the session ended: start over at sign-in.
addEventListener('lldprep:signed-out', () => { if (AUTH.mode === 'accounts') location.reload(); });

boot().catch(bootFailed);

export { itemKey };
