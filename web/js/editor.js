// Monaco is loaded lazily through its AMD loader (vendored, offline).
import { h } from 'preact';
import { useEffect, useRef } from 'preact/hooks';

let monacoPromise = null;

export function loadMonaco() {
  if (monacoPromise) return monacoPromise;
  monacoPromise = new Promise((resolve, reject) => {
    const start = () => {
      window.MonacoEnvironment = {
        getWorkerUrl: () => new URL('vendor/monaco/vs/base/worker/workerMain.js', location.href).href,
      };
      window.require.config({ paths: { vs: 'vendor/monaco/vs' } });
      window.require(['vs/editor/editor.main'], () => {
        applyTheme();
        resolve(window.monaco);
      }, reject);
    };
    if (window.require && window.require.config) start();
    else window.addEventListener('load', start, { once: true });
  });
  return monacoPromise;
}

export function isDark() {
  const t = document.documentElement.dataset.theme;
  return t ? t === 'dark' : matchMedia('(prefers-color-scheme: dark)').matches;
}

export function applyTheme() {
  if (window.monaco) window.monaco.editor.setTheme(isDark() ? 'vs-dark' : 'vs');
}

// Follow OS theme changes while the app is open (unless the user picked one).
matchMedia('(prefers-color-scheme: dark)').addEventListener('change', applyTheme);

/**
 * Code editor component.
 * props: value, lang ('java'|'go'), onChange(code), onRun(), onSubmit(), readOnly, height
 */
export function CodeEditor({ value, lang, onChange, onRun, onSubmit, readOnly = false }) {
  const host = useRef(null);
  const ed = useRef(null);
  const cbs = useRef({});
  cbs.current = { onChange, onRun, onSubmit };

  useEffect(() => {
    let disposed = false;
    loadMonaco().then((monaco) => {
      if (disposed || !host.current) return;
      ed.current = monaco.editor.create(host.current, {
        value: value || '',
        language: lang,
        automaticLayout: true,
        minimap: { enabled: false },
        fontSize: 13.5,
        fontFamily: "'JetBrains Mono', 'SF Mono', Menlo, Consolas, monospace",
        tabSize: 4,
        insertSpaces: lang !== 'go',
        scrollBeyondLastLine: false,
        readOnly,
        renderLineHighlight: 'line',
        padding: { top: 12 },
        fixedOverflowWidgets: true,
      });
      ed.current.onDidChangeModelContent(() => cbs.current.onChange?.(ed.current.getValue()));
      ed.current.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Quote, () => cbs.current.onRun?.());
      ed.current.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Enter, () => cbs.current.onSubmit?.());
      ed.current.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {});
    });
    return () => {
      disposed = true;
      ed.current?.getModel()?.dispose();
      ed.current?.dispose();
      ed.current = null;
    };
  }, []);

  useEffect(() => {
    const e = ed.current;
    if (!e) return;
    if (e.getValue() !== value) e.setValue(value || '');
  }, [value]);

  useEffect(() => {
    const e = ed.current;
    if (!e || !window.monaco) return;
    window.monaco.editor.setModelLanguage(e.getModel(), lang);
    e.updateOptions({ insertSpaces: lang !== 'go' });
  }, [lang]);

  return h('div', { class: 'editor-host', ref: host });
}
