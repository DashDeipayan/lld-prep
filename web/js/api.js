async function req(method, url, body, raw) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.body = raw ? body : JSON.stringify(body);
    opts.headers['Content-Type'] = raw ? 'text/plain' : 'application/json';
  }
  const r = await fetch(url, opts);
  if (!r.ok) throw new Error(`${method} ${url}: ${r.status} ${await r.text()}`);
  if (r.status === 204) return null;
  return r.json();
}

const cache = new Map();
function cached(url) {
  if (!cache.has(url)) cache.set(url, req('GET', url).catch((e) => { cache.delete(url); throw e; }));
  return cache.get(url);
}

export const api = {
  plan: () => cached('/api/plan'),
  lesson: (id) => cached(`/api/lesson/${encodeURIComponent(id)}`),
  exercise: (id) => cached(`/api/exercise/${encodeURIComponent(id)}`),
  tutorial: (id) => cached(`/api/tutorial/${encodeURIComponent(id)}`),
  quiz: (id) => cached(`/api/quiz/${encodeURIComponent(id)}`),
  solution: (kind, id, lang) => req('GET', `/api/solution/${kind}/${encodeURIComponent(id)}/${lang}`),
  search: (q) => req('GET', `/api/search?q=${encodeURIComponent(q)}`),
  toolchains: () => cached('/api/toolchains'),
  state: () => req('GET', '/api/state'),
  saveState: (s) => req('PUT', '/api/state', s),
  draft: (key) => req('GET', `/api/draft/${encodeURIComponent(key)}`),
  saveDraft: (key, code) => req('PUT', `/api/draft/${encodeURIComponent(key)}`, code, true),
  deleteDraft: (key) => req('DELETE', `/api/draft/${encodeURIComponent(key)}`),
  run: (body) => req('POST', '/api/run', body),
};
