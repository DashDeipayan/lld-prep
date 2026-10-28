import { useState, useEffect } from 'preact/hooks';
import { api } from './api.js';
import { update, markDone } from './state.js';
import { html, useAppState, useAsync, Markdown, BookmarkButton, Loading, ErrorBox, Icon } from './ui.js';
import { NextUp, Crumbs } from './app.js';

const PASS = 0.7;

export function QuizView({ id }) {
  const r = useAsync(() => api.quiz(id), [id]);
  const s = useAppState();
  const [answers, setAnswers] = useState({});
  const [submitted, setSubmitted] = useState(false);
  useEffect(() => { window.scrollTo(0, 0); }, [submitted]);
  if (r.loading) return html`<${Loading} />`;
  if (r.error) return html`<${ErrorBox} error=${r.error} />`;
  const q = r.data;
  const total = q.questions.length;
  const answered = Object.keys(answers).length;
  const score = q.questions.reduce((a, qq, i) => a + (answers[i] === qq.answer ? 1 : 0), 0);
  const prev = s.quiz[id];

  const submit = () => {
    setSubmitted(true);
    update((st) => {
      const old = st.quiz[id];
      st.quiz[id] = { best: Math.max(old?.best || 0, score), last: score, total, at: new Date().toISOString(), attempts: (old?.attempts || 0) + 1 };
    });
    if (score / total >= PASS) markDone('quiz', id, true);
  };
  const retry = () => { setAnswers({}); setSubmitted(false); };

  return html`<div class="page narrow quiz">
    <${Crumbs} type="quiz" id=${id} />
    <div class="title-row"><h1>${q.title}</h1><${BookmarkButton} bkey=${'quiz:' + id} type="quiz" id=${id} title=${q.title} compact /></div>
    ${q.summary ? html`<p class="lead">${q.summary}</p>` : null}
    <div class="meta-row"><span>${total} questions</span><span>Pass mark ${Math.round(PASS * 100)}%</span>
      ${prev ? html`<span>Best so far: ${prev.best}/${prev.total}</span>` : null}</div>

    ${submitted ? html`<div class=${'score-card ' + (score / total >= PASS ? 'ok' : 'bad')}>
      <div class="score-big">${score}<span>/${total}</span></div>
      <div><b>${score / total >= PASS ? 'Passed.' : 'Not yet.'}</b> ${score / total >= PASS
        ? 'Read the explanations for anything you got wrong, then move on.'
        : 'Re-read the lesson sections behind the wrong answers, then try again.'}</div>
      <button class="btn" onClick=${retry}><${Icon} name="reset" size=${15} /> Retake</button>
    </div>` : null}

    <ol class="questions">
      ${q.questions.map((qq, i) => {
        const chosen = answers[i];
        return html`<li class="question">
          <${Markdown} src=${qq.q} class="q-text" />
          ${qq.code ? html`<${Markdown} src=${qq.code} />` : null}
          <div class="options" role="radiogroup">
            ${qq.options.map((o, j) => {
              let cls = 'option';
              if (chosen === j) cls += ' chosen';
              if (submitted && j === qq.answer) cls += ' correct';
              if (submitted && chosen === j && j !== qq.answer) cls += ' wrong';
              return html`<button type="button" role="radio" aria-checked=${chosen === j} class=${cls} disabled=${submitted}
                onClick=${() => setAnswers({ ...answers, [i]: j })}>
                <span class="opt-letter">${String.fromCharCode(65 + j)}</span><${Markdown} src=${o} class="opt-text" /></button>`;
            })}
          </div>
          ${submitted ? html`<div class=${'explain ' + (chosen === qq.answer ? 'ok' : 'bad')}>
            <b>${chosen === qq.answer ? 'Correct.' : chosen == null ? 'Not answered.' : 'Incorrect.'}</b> <${Markdown} src=${qq.explain} /></div>` : null}
        </li>`;
      })}
    </ol>
    ${!submitted ? html`<div class="quiz-submit">
      <span class="muted">${answered}/${total} answered</span>
      <button class="btn primary lg" disabled=${answered === 0} onClick=${submit}>Submit answers</button>
    </div>` : html`<${NextUp} type="quiz" id=${id} onDone=${false} />`}
  </div>`;
}
