// Exercise and build-along (tutorial) views, sharing one editor workbench.
import { useEffect, useRef, useState, useCallback } from 'preact/hooks';
import { api } from './api.js';
import { update, markDone, getState } from './state.js';
import { html, useAppState, useAsync, Icon, Markdown, BookmarkButton, LangToggle, Loading, ErrorBox, Difficulty, setLang, fmtMin, useSectionBookmarks, scrollToAnchor } from './ui.js';
import { CodeEditor } from './editor.js';
import { LANGS, langLabel, fileName } from './langs.js';
import { NextUp, Crumbs } from './app.js';

export function humanizeTest(name) {
  let step = null;
  let n = name.replace(/^Test/, '').replace(/^test/, '');
  const m = n.match(/^[sS](\d+)_(.*)$/);
  if (m) { step = Number(m[1]); n = m[2]; }
  n = n.replace(/_/g, ' ').replace(/([a-z0-9])([A-Z])/g, '$1 $2').toLowerCase().trim();
  n = n.charAt(0).toUpperCase() + n.slice(1);
  return step ? `Step ${step} · ${n}` : n;
}

const STATUS = {
  ok: ['ok', 'All good'],
  failed: ['bad', 'Some tests failed'],
  compile_error: ['bad', 'Compilation failed'],
  runtime_error: ['bad', 'Runtime error'],
  timeout: ['bad', 'Timed out'],
  toolchain_missing: ['warn', 'Toolchain missing'],
};

export function Results({ result, running, mode }) {
  if (running) return html`<div class="results"><div class="results-head running"><div class="spinner" /> ${mode === 'run' ? 'Compiling and running…' : 'Running tests…'}</div></div>`;
  if (!result) return html`<div class="results empty-results"><span class="muted"><kbd>⌘ '</kbd> runs <code>main</code>. <kbd>⌘ Enter</kbd> runs the tests.</span></div>`;
  const [cls, label] = STATUS[result.status] || ['bad', result.status];
  const isTest = result.total > 0 || mode === 'test';
  return html`<div class="results">
    <div class=${'results-head ' + cls}>
      <b>${result.status === 'ok' && isTest ? `All ${result.total} tests passed` : result.status === 'ok' ? 'Ran successfully' : label}</b>
      ${result.total ? html`<span>${result.passed}/${result.total} passing</span>` : null}
      <span class="muted small">${(result.durationMs / 1000).toFixed(1)}s</span>
    </div>
    ${result.tests?.length ? html`<ul class="test-list">${result.tests.map((t) => html`<li class=${t.passed ? 'pass' : 'fail'}>
      <span class="ti">${t.passed ? '✓' : '✗'}</span>
      <div><div class="tn">${humanizeTest(t.name)}</div>${!t.passed && t.message ? html`<pre class="tm">${t.message}</pre>` : null}</div>
    </li>`)}</ul>` : null}
    ${result.stdout?.trim() ? html`<div class="out"><div class="out-label">stdout</div><pre>${result.stdout}</pre></div>` : null}
    ${result.stderr?.trim() ? html`<div class="out err"><div class="out-label">${result.status === 'compile_error' ? 'compiler' : 'stderr'}</div><pre>${result.stderr}</pre></div>` : null}
  </div>`;
}

/**
 * Editor + toolbar + results.
 * draftKey(lang) → storage key; starter: {java, go, python}; runBody(lang, code, mode) → request; onResult(result, mode, lang)
 */
export function Workbench({ draftKey, starter, runBody, onResult, header, testLabel = 'Run tests' }) {
  const s = useAppState();
  const lang = s.lang;
  const [code, setCode] = useState(null);
  const [result, setResult] = useState(null);
  const [running, setRunning] = useState(null);
  const saveT = useRef(null);
  const codeRef = useRef('');
  const available = starter[lang] != null;

  useEffect(() => {
    let live = true;
    setCode(null);
    setResult(null);
    if (!available) return;
    api.draft(draftKey(lang)).then((d) => {
      if (!live) return;
      const c = d.exists ? d.code : starter[lang];
      codeRef.current = c;
      setCode(c);
    }).catch(() => { codeRef.current = starter[lang]; setCode(starter[lang]); });
    return () => { live = false; };
  }, [lang, draftKey(lang)]);

  const onChange = (c) => {
    codeRef.current = c;
    clearTimeout(saveT.current);
    const key = draftKey(lang);
    saveT.current = setTimeout(() => api.saveDraft(key, c).catch(console.error), 600);
  };

  const run = async (mode) => {
    if (running) return;
    setRunning(mode);
    try {
      clearTimeout(saveT.current);
      api.saveDraft(draftKey(lang), codeRef.current).catch(() => {});
      const r = await api.run(runBody(lang, codeRef.current, mode));
      setResult(r);
      onResult?.(r, mode, lang);
    } catch (e) {
      setResult({ status: 'runtime_error', stderr: String(e.message || e), durationMs: 0 });
    } finally {
      setRunning(null);
    }
  };

  const reset = () => {
    if (!confirm(`Replace your ${langLabel(lang)} code with the original starter? This cannot be undone.`)) return;
    api.deleteDraft(draftKey(lang)).catch(() => {});
    codeRef.current = starter[lang];
    setCode(starter[lang]);
    setResult(null);
  };

  const loadCode = (c) => {
    codeRef.current = c;
    setCode(c);
    api.saveDraft(draftKey(lang), c).catch(() => {});
  };
  Workbench.load = loadCode;

  return html`<div class="workbench">
    <div class="wb-toolbar">
      <${LangToggle} lang=${lang} onChange=${setLang} />
      <span class="file-name">${fileName(lang)}</span>
      ${header}
      <div class="spacer" />
      <button class="btn ghost" onClick=${reset} disabled=${!available} title="Reset to starter code"><${Icon} name="reset" size=${15} /> Reset</button>
      <button class="btn" onClick=${() => run('run')} disabled=${!!running || !available} title="Run main (⌘ ')"><${Icon} name="play" size=${14} /> Run</button>
      <button class="btn primary" onClick=${() => run('test')} disabled=${!!running || !available} title="Run tests (⌘ Enter)"><${Icon} name="flask" size=${15} /> ${testLabel}</button>
    </div>
    <div class="wb-editor">
      ${!available ? html`<div class="unavailable"><p>This one isn't available in ${langLabel(lang)} yet.</p>
          <p class="muted">Switch to ${LANGS.filter((l) => starter[l] != null).map(langLabel).join(' or ')} above.</p></div>`
        : code == null ? html`<${Loading} />` : html`<${CodeEditor} value=${code} lang=${lang} onChange=${onChange}
          onRun=${() => run('run')} onSubmit=${() => run('test')} />`}
    </div>
    <${Results} result=${result} running=${!!running} mode=${running || (result && result.total ? 'test' : 'run')} />
  </div>`;
}

function SolutionPanel({ kind, id, lang, intro }) {
  const [shown, setShown] = useState(false);
  const r = useAsync(() => (shown ? api.solution(kind, id, lang) : Promise.resolve(null)), [shown, lang]);
  if (!shown) {
    return html`<div class="reveal">
      <${Icon} name="eye" size=${28} />
      <p>${intro || 'Try to get every test green first. Interviewers can tell a memorised design from one you reasoned through.'}</p>
      <button class="btn" onClick=${() => setShown(true)}>Reveal reference solution</button>
    </div>`;
  }
  if (r.loading || !r.data) return html`<${Loading} />`;
  return html`<div>
    <div class="btn-row"><button class="btn" onClick=${() => { if (confirm('Load the reference solution into the editor? Your current code will be replaced (use Reset afterwards to get the starter back).')) Workbench.load?.(r.data.code); }}>Load into editor</button></div>
    <${Markdown} src=${'```' + lang + '\n' + r.data.code + '\n```'} />
  </div>`;
}

function Split({ left, right }) {
  return html`<div class="split"><div class="split-left">${left}</div><div class="split-right">${right}</div></div>`;
}

export function ExerciseView({ id }) {
  const r = useAsync(() => api.exercise(id), [id]);
  const s = useAppState();
  const [tab, setTab] = useState('problem');
  const [hints, setHints] = useState(0);
  const draftKey = useCallback((lang) => `exercise:${id}:${lang}`, [id]);
  if (r.loading) return html`<${Loading} />`;
  if (r.error) return html`<${ErrorBox} error=${r.error} />`;
  const ex = r.data;
  const passedIn = LANGS.filter((l) => s.passed[`exercise:${id}:${l}`]);
  const onResult = (res, mode, lang) => {
    if (mode === 'test' && res.status === 'ok' && res.total > 0) {
      update((st) => { st.passed[`exercise:${id}:${lang}`] = st.passed[`exercise:${id}:${lang}`] || new Date().toISOString(); });
      markDone('exercise', id, true);
    }
  };
  const left = html`<div class="panel">
    <${Crumbs} type="exercise" id=${id} />
    <div class="title-row"><h1>${ex.title}</h1><${BookmarkButton} bkey=${'exercise:' + id} type="exercise" id=${id} title=${ex.title} compact /></div>
    <div class="meta-row"><${Difficulty} level=${ex.difficulty} /><span>${fmtMin(ex.minutes)}</span>
      ${passedIn.map((l) => html`<span class="chip ok">✓ ${langLabel(l)}</span>`)}</div>
    <div class="tabs small">
      <button class=${tab === 'problem' ? 'on' : ''} onClick=${() => setTab('problem')}>Problem</button>
      <button class=${tab === 'hints' ? 'on' : ''} onClick=${() => setTab('hints')}>Hints <span class="count">${ex.hints.length}</span></button>
      <button class=${tab === 'solution' ? 'on' : ''} onClick=${() => setTab('solution')}>Solution</button>
    </div>
    ${tab === 'problem' ? html`<${Markdown} src=${ex.problem} />` : null}
    ${tab === 'hints' ? html`<div class="hints">
      ${ex.hints.slice(0, hints).map((h, i) => html`<div class="hint"><b>Hint ${i + 1}</b><${Markdown} src=${h} /></div>`)}
      ${hints < ex.hints.length ? html`<button class="btn" onClick=${() => setHints(hints + 1)}><${Icon} name="bulb" size=${15} /> Show hint ${hints + 1} of ${ex.hints.length}</button>` : html`<p class="muted">That's all the hints.</p>`}
    </div>` : null}
    ${tab === 'solution' ? html`<${SolutionPanel} kind="exercise" id=${id} lang=${s.lang} />` : null}
    <${NextUp} type="exercise" id=${id} />
  </div>`;
  const right = html`<${Workbench} draftKey=${draftKey} starter=${ex.starter}
    runBody=${(lang, code, mode) => ({ kind: 'exercise', id, lang, code, mode })} onResult=${onResult} />`;
  return html`<${Split} left=${left} right=${right} />`;
}

export function TutorialView({ id, anchor }) {
  const r = useAsync(() => api.tutorial(id), [id]);
  const s = useAppState();
  const [step, setStep] = useState(0); // 0 = intro
  const draftKey = useCallback((lang) => `tutorial:${id}:${lang}`, [id]);
  const scroller = useRef(null);
  const docRef = useRef(null);
  useSectionBookmarks(docRef, { type: 'tutorial', id, title: r.data?.title }, [r.data, step]);
  useEffect(() => { scroller.current?.scrollTo(0, 0); }, [step]);
  useEffect(() => { if (r.data) scrollToAnchor(anchor); }, [r.data]);
  if (r.loading) return html`<${Loading} />`;
  if (r.error) return html`<${ErrorBox} error=${r.error} />`;
  const t = r.data;
  const lang = s.lang;
  const passed = s.tutorialStep[`${id}:${lang}`] || 0;
  const n = t.steps.length;
  const stepTarget = step === 0 ? Math.min(passed + 1, n) : step;
  const onResult = (res, mode) => {
    if (mode !== 'test' || res.status !== 'ok' || res.total === 0) return;
    update((st) => { st.tutorialStep[`${id}:${lang}`] = Math.max(st.tutorialStep[`${id}:${lang}`] || 0, stepTarget); });
    if (stepTarget === n) {
      update((st) => { st.passed[`tutorial:${id}:${lang}`] = st.passed[`tutorial:${id}:${lang}`] || new Date().toISOString(); });
      markDone('tutorial', id, true);
    }
  };
  const cur = step === 0 ? null : t.steps[step - 1];
  const left = html`<div class="panel" ref=${scroller}>
    <${Crumbs} type="tutorial" id=${id} />
    <div class="title-row"><h1>${t.title}</h1><${BookmarkButton} bkey=${'tutorial:' + id} type="tutorial" id=${id} title=${t.title} compact /></div>
    <div class="stepper">
      <button class=${'step-dot' + (step === 0 ? ' on' : '')} onClick=${() => setStep(0)} title="Overview">i</button>
      ${t.steps.map((st, i) => {
        const k = i + 1;
        const locked = k > passed + 1;
        return html`<button class=${'step-dot' + (step === k ? ' on' : '') + (k <= passed ? ' passed' : '') + (locked ? ' locked' : '')}
          disabled=${locked} title=${locked ? 'Pass the previous step first' : st.title} onClick=${() => setStep(k)}>${k <= passed ? '✓' : k}</button>`;
      })}
      <span class="muted small">${passed}/${n} steps passed in ${langLabel(lang)}</span>
    </div>
    <div ref=${docRef}>
      ${step === 0 ? html`<${Markdown} src=${t.intro} />` : html`<div class="eyebrow">Step ${step} of ${n}</div><h2 class="step-title">${cur.title}</h2><${Markdown} src=${cur.markdown} />`}
    </div>
    <div class="step-nav">
      ${step > 0 ? html`<button class="btn" onClick=${() => setStep(step - 1)}><${Icon} name="left" size=${15} /> Back</button>` : html`<span />`}
      ${step === 0 ? html`<button class="btn primary" onClick=${() => setStep(Math.min(passed + 1, n))}>${passed ? `Resume at step ${Math.min(passed + 1, n)}` : 'Start step 1'} <${Icon} name="chevron" size=${15} /></button>`
        : step < n ? html`<button class=${'btn' + (step <= passed ? ' primary' : '')} disabled=${step > passed} title=${step > passed ? 'Make this step’s tests pass to continue' : ''}
            onClick=${() => setStep(step + 1)}>${step > passed ? html`<${Icon} name="lock" size=${14} /> ` : null}Next step <${Icon} name="chevron" size=${15} /></button>`
        : passed >= n ? html`<span class="chip ok">Build-along complete 🎉</span>` : null}
    </div>
    ${step === n || passed >= n ? html`<details class="ref-details"><summary>Final reference solution</summary><${SolutionPanel} kind="tutorial" id=${id} lang=${lang} intro="The complete solution after the last step." /></details>` : null}
    <${NextUp} type="tutorial" id=${id} onDone=${false} />
  </div>`;
  const right = html`<${Workbench} draftKey=${draftKey} starter=${t.starter} testLabel=${`Check step ${stepTarget}`}
    runBody=${(lang, code, mode) => ({ kind: 'tutorial', id, lang, code, mode, step: stepTarget })} onResult=${onResult} />`;
  return html`<${Split} left=${left} right=${right} />`;
}
