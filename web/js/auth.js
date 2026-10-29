// Sign-in / sign-up page, shown only when the server runs in accounts mode
// (started with a MongoDB URI) and nobody is signed in.
import { useState } from 'preact/hooks';
import { api } from './api.js';
import { html } from './ui.js';

export function SignIn({ inviteRequired, onSignedIn }) {
  const [mode, setMode] = useState('login');
  const [form, setForm] = useState({ username: '', password: '', confirm: '', invite: '' });
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const signup = mode === 'signup';
  const set = (k) => (e) => setForm({ ...form, [k]: e.target.value });

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

  const switchMode = () => { setMode(signup ? 'login' : 'signup'); setError(''); };

  return html`<div class="auth-page">
    <form class="card auth-card" onSubmit=${submit}>
      <div class="auth-brand">
        <span class="logo"><svg viewBox="0 0 32 32" width="30" height="30"><rect width="32" height="32" rx="7" fill="currentColor"/><path d="M9 9h5v5H9zM18 9h5v5h-5zM13.5 18h5v5h-5z" fill="var(--bg)"/></svg></span>
        <div><b>LLD Prep</b><small>14 days to the offer</small></div>
      </div>
      <h1>${signup ? 'Create your account' : 'Sign in'}</h1>
      <p class="muted">${signup ? 'Your progress, quiz scores, bookmarks and code are saved to your account.' : 'Pick up where you left off.'}</p>
      <label class="field">Username
        <input name="username" autocomplete="username" required minlength="3" maxlength="32" autofocus
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
        ${inviteRequired ? html`<label class="field">Invite code
          <input name="invite" autocomplete="off" required value=${form.invite} onInput=${set('invite')} />
        </label>` : null}
        <p class="muted small">3–32 characters: letters, digits, <code>.</code> <code>_</code> <code>-</code>. Password at least 8 characters.</p>` : null}
      ${error ? html`<p class="auth-error" role="alert">${error}</p>` : null}
      <button class="btn primary lg auth-submit" disabled=${busy}>${busy ? 'Please wait…' : signup ? 'Create account' : 'Sign in'}</button>
      <p class="auth-switch muted">
        ${signup ? 'Already have an account? ' : 'New here? '}
        <a href="#" onClick=${(e) => { e.preventDefault(); switchMode(); }}>${signup ? 'Sign in' : 'Create an account'}</a>
      </p>
    </form>
  </div>`;
}
