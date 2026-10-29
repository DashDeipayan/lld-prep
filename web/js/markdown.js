// Markdown → HTML with LLD-specific extensions:
//  * adjacent ```java / ```go / ```python fences become one tabbed block (follows the global
//    language; a block without the current language shows its first tab)
//  * ```mermaid fences render as diagrams
//  * <div class="lang-note" data-lang="go"> blocks show only for that language
//  * h2/h3 get ids so sections can be linked and bookmarked
import { marked } from 'marked';
import { loadMonaco } from './editor.js';
import { LANGS, langLabel } from './langs.js';

marked.use({ gfm: true });

export const slug = (s) => s.toLowerCase().replace(/<[^>]+>/g, '').replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');

const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

export function renderMarkdown(src) {
  const tokens = marked.lexer(src);
  const out = [];
  for (let i = 0; i < tokens.length; i++) {
    const t = tokens[i];
    if (t.type === 'code' && LANGS.includes(t.lang)) {
      // Collect a run of java/go/python fences separated only by whitespace.
      const group = [t];
      let j = i + 1;
      while (j < tokens.length) {
        if (tokens[j].type === 'space') { j++; continue; }
        if (tokens[j].type === 'code' && LANGS.includes(tokens[j].lang) && !group.some((g) => g.lang === tokens[j].lang)) {
          group.push(tokens[j]);
          j++;
          continue;
        }
        break;
      }
      i = j - 1;
      out.push({ type: 'html', raw: codeGroup(group), text: codeGroup(group), block: true });
      continue;
    }
    if (t.type === 'code' && t.lang === 'mermaid') {
      const html = `<div class="mermaid-wrap"><pre class="mermaid">${esc(t.text)}</pre></div>`;
      out.push({ type: 'html', raw: html, text: html, block: true });
      continue;
    }
    if (t.type === 'heading' && (t.depth === 2 || t.depth === 3)) {
      const html = `<h${t.depth} id="${slug(t.text)}" class="anchored">${marked.parseInline(t.text)}</h${t.depth}>`;
      out.push({ type: 'html', raw: html, text: html, block: true });
      continue;
    }
    out.push(t);
  }
  out.links = tokens.links;
  return marked.parser(out);
}

function codeGroup(group) {
  const tabs = group.length > 1
    ? `<div class="code-tabs">${group.map((g) => `<button type="button" data-lang-tab="${g.lang}">${langLabel(g.lang)}</button>`).join('')}</div>`
    : `<div class="code-tabs single"><span>${langLabel(group[0].lang)}</span></div>`;
  const bodies = group.map((g) => `<pre class="code" data-lang-body="${g.lang}"><code data-lang="${g.lang}">${esc(g.text)}</code></pre>`).join('');
  // "no-python" etc. lets CSS fall back to the first tab when the chosen language is missing.
  const missing = LANGS.filter((l) => !group.some((g) => g.lang === l)).map((l) => ' no-' + l).join('');
  return `<div class="code-group${group.length > 1 ? ' multi' + missing : ''}">${tabs}<button type="button" class="copy-btn" title="Copy">Copy</button>${bodies}</div>`;
}

let mermaidReady = false;
function initMermaid() {
  if (!window.mermaid) return false;
  if (!mermaidReady) {
    const dark = document.documentElement.dataset.theme === 'dark' ||
      (!document.documentElement.dataset.theme && matchMedia('(prefers-color-scheme: dark)').matches);
    window.mermaid.initialize({ startOnLoad: false, theme: dark ? 'dark' : 'neutral', securityLevel: 'strict', fontFamily: 'inherit' });
    mermaidReady = true;
  }
  return true;
}

/** Post-process a container after its markdown HTML was inserted. */
export async function enhance(root) {
  if (!root) return;
  root.querySelectorAll('.copy-btn').forEach((b) => {
    b.onclick = () => {
      const visible = [...b.parentElement.querySelectorAll('pre.code')].find((p) => p.offsetParent !== null) || b.parentElement.querySelector('pre.code');
      navigator.clipboard?.writeText(visible.dataset.raw || visible.innerText);
      b.textContent = 'Copied';
      setTimeout(() => (b.textContent = 'Copy'), 1200);
    };
  });
  const nodes = [...root.querySelectorAll('pre.mermaid:not([data-processed])')];
  if (nodes.length && initMermaid()) {
    try { await window.mermaid.run({ nodes }); } catch (e) { console.warn('mermaid', e); }
  }
  const codes = [...root.querySelectorAll('code[data-lang]:not([data-colored])')];
  if (codes.length) {
    const monaco = await loadMonaco();
    for (const c of codes) {
      c.dataset.colored = '1';
      c.parentElement.dataset.raw = c.textContent;
      try {
        c.innerHTML = await monaco.editor.colorize(c.textContent, c.dataset.lang, { tabSize: 4 });
      } catch (e) { /* keep plain text */ }
    }
  }
}

export function resetMermaidTheme() { mermaidReady = false; }
