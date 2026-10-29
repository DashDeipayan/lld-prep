// Sign-in / sign-up page, shown only when the server runs in accounts mode
// (started with a MongoDB URI) and nobody is signed in.
import { useEffect, useState } from 'preact/hooks';
import { api } from './api.js';
import { html } from './ui.js';

export const PROVIDER_LABEL = { google: 'Google', github: 'GitHub' };

// Official marks: Google's "G" and the GitHub mark (Simple Icons, CC0).
const GOOGLE_G = html`<svg viewBox="0 0 48 48" width="18" height="18" aria-hidden="true"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/></svg>`;
const GITHUB_MARK = html`<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" fill="currentColor"><path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/></svg>`;
export const PROVIDER_ICON = { google: GOOGLE_G, github: GITHUB_MARK };

/** Full-page navigation to the provider; the server redirects back to the app. */
export function startProvider(provider, params = {}) {
  const q = new URLSearchParams(params).toString();
  location.href = `/api/auth/${provider}/start${q ? '?' + q : ''}`;
}

/**
 * Reads a one-off value the server put in the hash (e.g. #/?auth_error=…) and
 * removes it from the address bar so a reload doesn't show it again.
 */
export function takeHashParam(name) {
  const [path, query = ''] = location.hash.split('?');
  const q = new URLSearchParams(query);
  const value = q.get(name);
  if (value != null) {
    q.delete(name);
    const rest = q.toString();
    history.replaceState(null, '', location.pathname + location.search + (path || '#/') + (rest ? '?' + rest : ''));
  }
  return value;
}

export function ProviderButton({ provider, onClick, label }) {
  return html`<button type="button" class=${'btn lg provider-btn provider-' + provider} onClick=${onClick}>
    ${PROVIDER_ICON[provider]}<span>${label || 'Continue with ' + PROVIDER_LABEL[provider]}</span>
  </button>`;
}

export function SignIn({ inviteRequired, providers = [], onSignedIn }) {
  const [mode, setMode] = useState('login');
  const [form, setForm] = useState({ username: '', password: '', confirm: '', invite: '' });
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const signup = mode === 'signup';
  const set = (k) => (e) => setForm({ ...form, [k]: e.target.value });

  useEffect(() => {
    const e = takeHashParam('auth_error');
    if (e) setError(e);
  }, []);

  const submit = async (e) => {
    e.preventDefault();
    setError('');
    if (signup && form.password !== form.confirm) {
      setError("Passwords don't match.");
      return;
    }
    setBusy(true);
    try {
      if (signup) await api.signup(form.username, form.password, form.invite);
      else await api.login(form.username, form.password);
      onSignedIn();
    } catch (err) {
      setError(err.message);
      setBusy(false);
    }
  };

  const continueWith = (provider) => {
    if (signup && inviteRequired && !form.invite.trim()) {
      setError('Enter your invite code first, then continue with ' + PROVIDER_LABEL[provider] + '.');
      return;
    }
    startProvider(provider, signup && form.invite.trim() ? { invite: form.invite.trim() } : {});
  };

  const switchMode = () => { setMode(signup ? 'login' : 'signup'); setError(''); };

  const inviteField = signup && inviteRequired ? html`<label class="field">Invite code
    <input name="invite" autocomplete="off" required value=${form.invite} onInput=${set('invite')} />
  </label>` : null;

  return html`<div class="auth-page">
    <form class="card auth-card" onSubmit=${submit}>
      <div class="auth-brand">
        <span class="logo"><svg viewBox="0 0 32 32" width="30" height="30"><rect width="32" height="32" rx="7" fill="currentColor"/><path d="M9 9h5v5H9zM18 9h5v5h-5zM13.5 18h5v5h-5z" fill="var(--bg)"/></svg></span>
        <div><b>LLD Prep</b><small>14 days to the offer</small></div>
      </div>
      <h1>${signup ? 'Create your account' : 'Sign in'}</h1>
      <p class="muted">${signup ? 'Your progress, quiz scores, bookmarks and code are saved to your account.' : 'Pick up where you left off.'}</p>
      ${error ? html`<p class="auth-error" role="alert">${error}</p>` : null}
      ${providers.length ? html`
        ${inviteField}
        <div class="provider-list">${providers.map((p) => html`<${ProviderButton} provider=${p} onClick=${() => continueWith(p)} />`)}</div>
        <div class="auth-divider"><span>or with a username</span></div>` : null}
      <label class="field">Username
        <input name="username" autocomplete="username" required minlength="3" maxlength="32" autofocus=${!providers.length}
          value=${form.username} onInput=${set('username')} />
      </label>
      <label class="field">Password
        <input name="password" type="password" autocomplete=${signup ? 'new-password' : 'current-password'} required minlength=${signup ? 8 : 1} maxlength="72"
          value=${form.password} onInput=${set('password')} />
      </label>
      ${signup ? html`
        <label class="field">Confirm password
          <input name="confirm" type="password" autocomplete="new-password" required value=${form.confirm} onInput=${set('confirm')} />
        </label>
        ${providers.length ? null : inviteField}
        <p class="muted small">3–32 characters: letters, digits, <code>.</code> <code>_</code> <code>-</code>. Password at least 8 characters.</p>` : null}
      <button class="btn primary lg auth-submit" disabled=${busy}>${busy ? 'Please wait…' : signup ? 'Create account' : 'Sign in'}</button>
      <p class="auth-switch muted">
        ${signup ? 'Already have an account? ' : 'New here? '}
        <a href="#" onClick=${(e) => { e.preventDefault(); switchMode(); }}>${signup ? 'Sign in' : 'Create an account'}</a>
      </p>
      <p class="auth-legal muted small">By continuing you agree to how we handle your data: <a href="/privacy">Privacy Policy</a>.</p>
    </form>
  </div>`;
}
