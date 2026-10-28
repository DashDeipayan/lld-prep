async function req(method, url, body, raw) {
  const opts = { method, headers: {} };
  if (body !== undefined) {
    opts.body = raw ? body : JSON.stringify(body);
    opts.headers['Content-Type'] = raw ? 'text/plain' : 'application/json';
  }
  const r = await fetch(url, opts);
  if (r.status === 401 && !url.startsWith('/api/auth/')) {
    // Session expired or signed out in another tab: show the sign-in page.
    dispatchEvent(new Event('lldprep:signed-out'));
  }
  if (!r.ok) {
    const err = new Error((await r.text()).trim() || `${method} ${url}: ${r.status}`);
    err.status = r.status;
    throw err;
  }
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
  me: () => req('GET', '/api/auth/me'),
  login: (username, password) => req('POST', '/api/auth/login', { username, password }),
  signup: (username, password, invite) => req('POST', '/api/auth/signup', { username, password, invite }),
  logout: () => req('POST', '/api/auth/logout'),
};
