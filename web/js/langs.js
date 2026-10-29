// The languages exercises and lesson code come in. Keep in sync with
// content.Langs on the server.
export const LANGS = ['java', 'go', 'python'];

const LABEL = { java: 'Java', go: 'Go', python: 'Python' };
const FILE = { java: 'Solution.java', go: 'solution.go', python: 'solution.py' };

export const langLabel = (l) => LABEL[l] || l;
export const fileName = (l) => FILE[l] || 'solution';
