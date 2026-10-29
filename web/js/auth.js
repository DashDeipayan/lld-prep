// Sign-in / sign-up page, shown only when the server runs in accounts mode
// (started with a MongoDB URI) and nobody is signed in.
import { useEffect, useState } from 'preact/hooks';
import { api } from './api.js';
import { html } from './ui.js';

export const PROVIDER_LABEL = { google: 'Google', apple: 'Apple' };

// Official marks: Google's "G" and Apple's logo (Simple Icons, CC0).
const GOOGLE_G = html`<svg viewBox="0 0 48 48" width="18" height="18" aria-hidden="true"><path fill="#EA4335" d="M24 9.5c3.54 0 6.71 1.22 9.21 3.6l6.85-6.85C35.9 2.38 30.47 0 24 0 14.62 0 6.51 5.38 2.56 13.22l7.98 6.19C12.43 13.72 17.74 9.5 24 9.5z"/><path fill="#4285F4" d="M46.98 24.55c0-1.57-.15-3.09-.38-4.55H24v9.02h12.94c-.58 2.96-2.26 5.48-4.78 7.18l7.73 6c4.51-4.18 7.09-10.36 7.09-17.65z"/><path fill="#FBBC05" d="M10.53 28.59c-.48-1.45-.76-2.99-.76-4.59s.27-3.14.76-4.59l-7.98-6.19C.92 16.46 0 20.12 0 24c0 3.88.92 7.54 2.56 10.78l7.97-6.19z"/><path fill="#34A853" d="M24 48c6.48 0 11.93-2.13 15.89-5.81l-7.73-6c-2.15 1.45-4.92 2.3-8.16 2.3-6.26 0-11.57-4.22-13.47-9.91l-7.98 6.19C6.51 42.62 14.62 48 24 48z"/></svg>`;
const APPLE_LOGO = html`<svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" fill="currentColor"><path d="M12.152 6.896c-.948 0-2.415-1.078-3.96-1.04-2.04.027-3.91 1.183-4.961 3.014-2.117 3.675-.546 9.103 1.519 12.09 1.013 1.454 2.208 3.09 3.792 3.039 1.52-.065 2.09-.987 3.935-.987 1.831 0 2.35.987 3.96.948 1.637-.026 2.676-1.48 3.676-2.948 1.156-1.688 1.636-3.325 1.662-3.415-.039-.013-3.182-1.221-3.22-4.857-.026-3.04 2.48-4.494 2.597-4.559-1.429-2.09-3.623-2.324-4.39-2.376-2-.156-3.675 1.09-4.61 1.09zM15.53 3.83c.843-1.012 1.4-2.427 1.245-3.83-1.207.052-2.662.805-3.532 1.818-.78.896-1.454 2.338-1.273 3.714 1.338.104 2.715-.688 3.559-1.701"/></svg>`;
export const PROVIDER_ICON = { google: GOOGLE_G, apple: APPLE_LOGO };

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
    </form>
  </div>`;
}
