// Design practice: guided, step-by-step design walkthroughs. Each step gives a
// task, the learner writes their own answer (saved as a draft), can open hints,
// ticks a self-check list, then reveals the model answer and compares.
import { useEffect, useRef, useState } from 'preact/hooks';
import { api } from './api.js';
import { update, markDone, isDone } from './state.js';
import { html, useAppState, useAsync, Icon, Markdown, BookmarkButton, Loading, ErrorBox, Difficulty, fmtMin, itemHref, TYPE_META } from './ui.js';
import { plan } from './app.js';

const draftKey = (id, step) => `practice:${id}:${step}`;

export function PracticeList() {
  useAppState();
  const list = plan().catalog.practice;
  return html`<div class="page">
    <h1>Design practice</h1>
    <p class="lead">Guided interview simulations. Each step asks you to do one part of the design yourself: requirements, entities, class diagram, core flows, patterns, concurrency and code. Then you compare with a model answer. Write before you peek; that's where the learning happens.</p>
    <div class="lib-grid">
      ${list.map((p) => {
        const done = isDone('practice', p.ID);
        return html`<a class=${'lib-card' + (done ? ' done' : '')} href=${itemHref('practice', p.ID)}>
          <div class="lib-card-top">
            <span class="type-badge t-practice"><${Icon} name="pencil" size=${14} />Design practice</span>
            ${done ? html`<span class="chip ok"><${Icon} name="check" size=${12} /> done</span>` : null}
          </div>
          <h3>${p.Title}</h3>
          ${p.Summary ? html`<p>${p.Summary}</p>` : null}
          <div class="lib-meta">
            <${Difficulty} level=${p.Difficulty} />
            ${p.Minutes ? html`<span>${fmtMin(p.Minutes)}</span>` : null}
            ${(p.Tags || []).slice(0, 3).map((t) => html`<span class="tag">${t}</span>`)}
          </div>
        </a>`;
      })}
    </div>
  </div>`;
}

/** The learner's own answer for one step, autosaved as a draft. */
function Notes({ id, step, onChange }) {
  const [text, setText] = useState(null);
  const t = useRef(null);
  useEffect(() => {
    let live = true;
    setText(null);
    api.draft(draftKey(id, step))
      .then((d) => { if (live) { setText(d.code || ''); onChange(d.code || ''); } })
      .catch(() => { if (live) { setText(''); onChange(''); } });
    return () => { live = false; clearTimeout(t.current); };
  }, [id, step]);
  if (text == null) return html`<${Loading} />`;
  const onInput = (e) => {
    const v = e.target.value;
    setText(v);
    onChange(v);
    clearTimeout(t.current);
    t.current = setTimeout(() => api.saveDraft(draftKey(id, step), v).catch(console.error), 600);
  };
  return html`<textarea class="practice-notes" value=${text} onInput=${onInput} spellcheck="false"
    placeholder="Write your answer here, the way you'd say it or sketch it on the whiteboard. Bullets are fine. Saved automatically." />`;
}

function Step({ p, n }) {
  const st = p.steps[n - 1];
  const [hints, setHints] = useState(0);
  const [revealed, setRevealed] = useState(false);
  const [checks, setChecks] = useState({});
  const [notes, setNotes] = useState('');
  const answerRef = useRef(null);
  const reveal = () => {
    setRevealed(true);
    setTimeout(() => answerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' }), 50);
  };
  return html`<div>
    <div class="eyebrow">Step ${n} of ${p.steps.length}</div>
    <h2 class="step-title">${st.title}</h2>
    <div class="practice-task">
      <div class="practice-label"><${Icon} name="pencil" size=${15} /> Your task</div>
      <${Markdown} src=${st.task} />
    </div>
    <${Notes} id=${p.id} step=${n} onChange=${setNotes} />
    ${st.hints.length ? html`<div class="hints practice-hints">
      ${st.hints.slice(0, hints).map((h, i) => html`<div class="hint"><b>Hint ${i + 1}</b><${Markdown} src=${h} /></div>`)}
      ${hints < st.hints.length ? html`<button class="btn ghost" onClick=${() => setHints(hints + 1)}><${Icon} name="bulb" size=${15} /> ${hints ? 'Another hint' : 'Stuck? Show a hint'} (${hints + 1}/${st.hints.length})</button>` : null}
    </div>` : null}
    ${st.checklist.length ? html`<div class="practice-checklist">
      <div class="practice-label"><${Icon} name="check" size=${15} /> Self-check: did your answer cover…</div>
      ${st.checklist.map((c, i) => html`<label class="check-item">
        <input type="checkbox" checked=${!!checks[i]} onChange=${(e) => setChecks({ ...checks, [i]: e.target.checked })} />
        <span>${c}</span></label>`)}
    </div>` : null}
    ${revealed ? html`<div class="practice-answer" ref=${answerRef}>
      <div class="practice-label"><${Icon} name="eye" size=${15} /> Model answer</div>
      <${Markdown} src=${st.answer} />
    </div>` : html`<div class="reveal practice-reveal">
      <p>${notes.trim() ? 'Done? Compare your answer with the model answer.' : 'Try it yourself first. Even three rough bullets beat reading the answer cold.'}</p>
      <button class=${'btn' + (notes.trim() ? ' primary' : '')} onClick=${reveal}><${Icon} name="eye" size=${15} /> ${notes.trim() ? 'Reveal model answer' : 'Reveal without attempting'}</button>
    </div>`}
  </div>`;
}

export function PracticeView({ id }) {
  const r = useAsync(() => api.practice(id), [id]);
  const s = useAppState();
  const [step, setStep] = useState(0); // 0 = brief
  useEffect(() => { window.scrollTo(0, 0); }, [step]);
  if (r.loading) return html`<${Loading} />`;
  if (r.error) return html`<${ErrorBox} error=${r.error} />`;
  const p = r.data;
  const n = p.steps.length;
  const reached = s.practiceStep[id] || 0;
  const complete = () => {
    update((st) => { st.practiceStep[id] = Math.max(st.practiceStep[id] || 0, step); });
    if (step === n) markDone('practice', id, true);
    else setStep(step + 1);
  };
  return html`<div class="page narrow practice">
    <div class="crumbs"><a href="#/practice">Design practice</a> / <span>${p.title}</span></div>
    <div class="title-row"><h1>${p.title}</h1><${BookmarkButton} bkey=${'practice:' + id} type="practice" id=${id} title=${p.title} compact /></div>
    <div class="meta-row"><${Difficulty} level=${p.difficulty} /><span>${fmtMin(p.minutes)}</span><span>${reached}/${n} steps done</span></div>
    <div class="stepper">
      <button class=${'step-dot' + (step === 0 ? ' on' : '')} onClick=${() => setStep(0)} title="Brief">i</button>
      ${p.steps.map((st, i) => {
        const k = i + 1;
        return html`<button class=${'step-dot' + (step === k ? ' on' : '') + (k <= reached ? ' passed' : '')}
          title=${st.title} onClick=${() => setStep(k)}>${k <= reached ? '✓' : k}</button>`;
      })}
    </div>
    ${step === 0 ? html`<div>
      <${Markdown} src=${p.intro} />
      <div class="callout tip"><b>How this works.</b> ${n} steps, in interview order. In each step you do the work in the box, use hints only if you're stuck, tick the self-check, then reveal the model answer and compare. Treat the gaps you find as your revision list.</div>
      <ol class="practice-outline">${p.steps.map((st, i) => html`<li><a href="javascript:void 0" onClick=${() => setStep(i + 1)}>${st.title}</a>${i < reached ? html` <span class="chip ok">✓</span>` : null}</li>`)}</ol>
      ${p.related?.length ? html`<p class="muted">Related: ${p.related.map((l, i) => html`${i ? ' · ' : ''}<a href=${itemHref(l.type, l.id)}>${TYPE_META[l.type]?.label}: ${l.title}</a>`)}</p>` : null}
    </div>` : html`<${Step} p=${p} n=${step} key=${step} />`}
    <div class="step-nav">
      ${step > 0 ? html`<button class="btn" onClick=${() => setStep(step - 1)}><${Icon} name="left" size=${15} /> Back</button>` : html`<span />`}
      ${step === 0 ? html`<button class="btn primary" onClick=${() => setStep(Math.min(reached + 1, n))}>${reached ? (reached >= n ? 'Review step 1' : `Resume at step ${reached + 1}`) : 'Start step 1'} <${Icon} name="chevron" size=${15} /></button>`
        : step < n ? html`<button class="btn primary" onClick=${complete}>${step <= reached ? 'Next step' : 'Mark done, next step'} <${Icon} name="chevron" size=${15} /></button>`
        : isDone('practice', id) ? html`<span class="chip ok">Practice complete 🎉</span>`
        : html`<button class="btn primary" onClick=${complete}><${Icon} name="check" size=${15} /> Finish practice</button>`}
    </div>
  </div>`;
}
