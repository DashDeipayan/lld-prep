// Timed mock interviews: brief → coding against the clock → tests + self-scored rubric.
import { useEffect, useState, useCallback } from 'preact/hooks';
import { api } from './api.js';
import { update, markDone, getState } from './state.js';
import { html, useAppState, useAsync, Icon, Markdown, Loading, ErrorBox, Difficulty, fmtMin, go } from './ui.js';
import { Workbench, Results } from './code.js';
import { plan, locate } from './app.js';

const SCALE = [
  [1, 'Weak'],
  [2, 'Mixed'],
  [3, 'Solid'],
  [4, 'Strong'],
];

function attemptScore(rubric, ratings) {
  const tw = rubric.reduce((a, r) => a + r.weight, 0);
  const got = rubric.reduce((a, r) => a + r.weight * ((ratings[r.id] || 1) - 1) / 3, 0);
  return Math.round((got / tw) * 100);
}

export function MocksView() {
  const s = useAppState();
  const P = plan();
  return html`<div class="page">
    <h1>Mock interviews</h1>
    <p class="lead">A full LLD round against the clock: clarify, design, code, test, then score yourself the way interviewers do. Hints and solutions stay hidden until you finish.</p>
    <div class="callout">
      <b>How to run a mock.</b> Talk out loud, or write your clarifying questions and assumptions as comments at the top of the file.
      Use the first 5–8 minutes on requirements and entities before you write code. Leave 5 minutes for extensions.
    </div>
    <div class="lib-grid">
      ${P.mocks.map((m) => {
        const attempts = s.mocks.filter((a) => a.id === m.id && a.score != null);
        const best = attempts.length ? Math.max(...attempts.map((a) => a.score)) : null;
        const loc = locate('mock', m.id);
        return html`<a class="lib-card" href=${'#/mock/' + m.id}>
          <div class="lib-card-top"><span class="type-badge t-mock"><${Icon} name="timer" size=${14} />Mock</span>
            ${best != null ? html`<span class=${'chip ' + (best >= 70 ? 'ok' : 'warn')}>best ${best}%</span>` : null}</div>
          <h3>${m.title}</h3>
          <p>${m.brief}</p>
          <div class="lib-meta"><${Difficulty} level=${m.difficulty} /><span>${fmtMin(m.minutes)}</span>
            ${m.company ? html`<span class="tag">${m.company}</span>` : null}${loc ? html`<span>Day ${loc.item.day}</span>` : null}
            <span>${attempts.length} attempt${attempts.length === 1 ? '' : 's'}</span></div>
        </a>`;
      })}
    </div>
    ${s.mocks.some((a) => a.score != null) ? html`<section class="card">
      <h2>History</h2>
      <table class="table"><thead><tr><th>Mock</th><th>Date</th><th>Lang</th><th>Time used</th><th>Tests</th><th>Score</th></tr></thead>
      <tbody>${s.mocks.filter((a) => a.score != null).slice().reverse().map((a) => html`<tr>
        <td>${P.mocks.find((m) => m.id === a.id)?.title || a.id}</td>
        <td>${new Date(a.startedAt).toLocaleString()}</td><td>${a.lang === 'go' ? 'Go' : 'Java'}</td>
        <td>${Math.round((a.endedAt - a.startedAt) / 60000)} min</td>
        <td>${a.tests ? `${a.tests.passed}/${a.tests.total}` : '–'}</td><td><b>${a.score}%</b></td></tr>`)}</tbody></table>
    </section>` : null}
  </div>`;
}

function useClock(active) {
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    if (!active) return;
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, [active]);
  return now;
}

const mmss = (ms) => {
  const neg = ms < 0;
  const t = Math.abs(Math.round(ms / 1000));
  return `${neg ? '+' : ''}${String(Math.floor(t / 60)).padStart(2, '0')}:${String(t % 60).padStart(2, '0')}`;
};

export function MockView({ id }) {
  const s = useAppState();
  const P = plan();
  const mock = P.mocks.find((m) => m.id === id);
  const ex = useAsync(() => (mock ? api.exercise(mock.exercise) : Promise.reject(new Error('Unknown mock'))), [id]);
  // The current attempt is the latest one for this mock without a score.
  const attempt = [...s.mocks].reverse().find((a) => a.id === id && a.score == null);
  const now = useClock(attempt && !attempt.endedAt);
  const draftKey = useCallback((lang) => `mock:${id}:${attempt?.startedAt || 0}:${lang}`, [id, attempt?.startedAt]);
  const [ratings, setRatings] = useState({});
  const [notes, setNotes] = useState('');
  const [finishing, setFinishing] = useState(false);

  if (!mock) return html`<div class="page"><${ErrorBox} error="Unknown mock" /></div>`;
  if (ex.loading) return html`<${Loading} />`;
  if (ex.error) return html`<${ErrorBox} error=${ex.error} />`;
  const e = ex.data;
  const patch = (fn) => update((st) => { const a = [...st.mocks].reverse().find((x) => x.id === id && x.score == null); if (a) fn(a); });

  const start = () => update((st) => { st.mocks.push({ id, lang: st.lang, startedAt: Date.now(), endedAt: null, score: null }); });

  if (!attempt) {
    const past = s.mocks.filter((a) => a.id === id && a.score != null);
    return html`<div class="page narrow">
      <div class="crumbs"><a href="#/mocks">Mock interviews</a> / ${mock.title}</div>
      <h1>${mock.title}</h1>
      <div class="meta-row"><${Difficulty} level=${e.difficulty} /><span>${fmtMin(mock.minutes)} on the clock</span>${mock.company ? html`<span>${mock.company}-style</span>` : null}</div>
      <p class="lead">${mock.brief}</p>
      <section class="card">
        <h2>Rules</h2>
        <ul>
          <li>The full problem statement appears when you start. Hints and the reference solution are hidden.</li>
          <li>The timer keeps running if you leave the page. Finish whenever you want, and overtime is recorded.</li>
          <li>When you finish, the hidden tests run and you score yourself on ${P.rubric.length} interviewer criteria.</li>
          <li>Language: <b>${s.lang === 'go' ? 'Go' : 'Java'}</b> (change it in the top bar before you start).</li>
        </ul>
        <button class="btn primary lg" onClick=${start}><${Icon} name="timer" size=${16} /> Start the clock</button>
      </section>
      ${past.length ? html`<p class="muted">Previous attempts: ${past.map((a) => a.score + '%').join(', ')}</p>` : null}
    </div>`;
  }

  const lang = attempt.lang;
  if (s.lang !== lang) setTimeout(() => update((st) => { st.lang = lang; }), 0);
  const deadline = attempt.startedAt + mock.minutes * 60000;
  const running = !attempt.endedAt;
  const left = deadline - (running ? now : attempt.endedAt);

  const finish = async () => {
    if (!confirm('Finish the interview and run the hidden tests?')) return;
    setFinishing(true);
    const endedAt = Date.now();
    let tests = null;
    let result = null;
    try {
      const d = await api.draft(draftKey(lang));
      const code = d.exists ? d.code : e.starter[lang];
      result = await api.run({ kind: 'exercise', id: mock.exercise, lang, code, mode: 'test' });
      tests = { passed: result.passed, total: result.total, status: result.status };
    } catch (err) {
      result = { status: 'runtime_error', stderr: String(err), durationMs: 0 };
    }
    patch((a) => { a.endedAt = endedAt; a.tests = tests; a.result = result; });
    setFinishing(false);
  };

  const abandon = () => {
    if (!confirm('Abandon this attempt? It will not be recorded.')) return;
    update((st) => { st.mocks = st.mocks.filter((a) => !(a.id === id && a.score == null)); });
  };

  const saveScore = () => {
    const score = attemptScore(P.rubric, ratings);
    patch((a) => { a.ratings = ratings; a.notes = notes; a.score = score; delete a.result; });
    if (locate('mock', id)) markDone('mock', id, true);
    go('mocks');
  };

  if (running) {
    const timer = html`<div class=${'timer' + (left < 0 ? ' over' : left < 5 * 60000 ? ' low' : '')}><${Icon} name="timer" size=${15} />${mmss(left)}</div>`;
    const leftPane = html`<div class="panel">
      <div class="crumbs"><a href="#/mocks">Mock interviews</a> / ${mock.title}</div>
      <div class="title-row"><h1>${e.title}</h1></div>
      <div class="callout tip">Write your clarifying questions and assumptions as comments first, then entities, then code.</div>
      <${Markdown} src=${e.problem} />
      <div class="btn-row"><button class="btn danger ghost" onClick=${abandon}>Abandon attempt</button></div>
    </div>`;
    const right = html`<${Workbench} draftKey=${draftKey} starter=${e.starter}
      header=${html`${timer}<button class="btn primary" disabled=${finishing} onClick=${finish}>${finishing ? 'Running tests…' : 'Finish interview'}</button>`}
      runBody=${(l, code, mode) => ({ kind: 'exercise', id: mock.exercise, lang: l, code, mode })} />`;
    return html`<div class="split mock-running"><div class="split-left">${leftPane}</div><div class="split-right locked-lang">${right}</div></div>`;
  }

  // Review phase: tests are in, now self-score.
  const used = Math.round((attempt.endedAt - attempt.startedAt) / 60000);
  const allRated = P.rubric.every((r) => ratings[r.id]);
  return html`<div class="page narrow">
    <div class="crumbs"><a href="#/mocks">Mock interviews</a> / ${mock.title}</div>
    <h1>Debrief: ${mock.title}</h1>
    <div class="meta-row"><span>Time used: <b>${used} min</b> of ${mock.minutes}</span>
      ${attempt.tests ? html`<span>Hidden tests: <b>${attempt.tests.passed}/${attempt.tests.total}</b></span>` : null}</div>
    ${attempt.result ? html`<${Results} result=${attempt.result} mode="test" />` : null}
    <section class="card">
      <h2>Score yourself like an interviewer would</h2>
      <p class="muted">Be strict. A “Strong” means you'd bet on getting a hire signal for that criterion.</p>
      <div class="rubric">
        ${P.rubric.map((r) => html`<div class="rubric-row">
          <div class="rubric-info"><b>${r.title}</b> <span class="muted small">weight ${r.weight}</span>
            <div class="small"><span class="ok-text">Strong:</span> ${r.strong}</div>
            <div class="small"><span class="bad-text">Weak:</span> ${r.weak}</div></div>
          <div class="seg rubric-seg">${SCALE.map(([v, label]) => html`<button class=${ratings[r.id] === v ? 'on' : ''}
            onClick=${() => setRatings({ ...ratings, [r.id]: v })}>${label}</button>`)}</div>
        </div>`)}
      </div>
      <label class="block-label" for="notes">Notes for next time</label>
      <textarea id="notes" rows="3" value=${notes} onInput=${(ev) => setNotes(ev.target.value)} placeholder="What would you do differently?" />
      <div class="btn-row">
        <button class="btn primary" disabled=${!allRated} onClick=${saveScore}>Save score${allRated ? ` (${attemptScore(P.rubric, ratings)}%)` : ''}</button>
        <a class="btn ghost" href=${'#/exercise/' + mock.exercise}>Open as a normal exercise (with solution)</a>
      </div>
    </section>
  </div>`;
}
