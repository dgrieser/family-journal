// Measure the caret position inside a <textarea> by rendering a hidden mirror
// <div> with the same typography and reading the offset of a marker span placed
// at the caret index. Returns coordinates relative to the textarea's border box.
//
// The mirror <div> and its marker <span> are created once and reused across
// calls; only the text content and copied computed styles are updated each
// time. This keeps scroll/input-driven recalculations cheap.

const MIRRORED_PROPERTIES: (keyof CSSStyleDeclaration)[] = [
  'direction',
  'boxSizing',
  'width',
  'height',
  'overflowX',
  'overflowY',
  'borderTopWidth',
  'borderRightWidth',
  'borderBottomWidth',
  'borderLeftWidth',
  'borderStyle',
  'paddingTop',
  'paddingRight',
  'paddingBottom',
  'paddingLeft',
  'fontStyle',
  'fontVariant',
  'fontWeight',
  'fontStretch',
  'fontSize',
  'fontSizeAdjust',
  'lineHeight',
  'fontFamily',
  'textAlign',
  'textTransform',
  'textIndent',
  'textDecoration',
  'letterSpacing',
  'wordSpacing',
  'tabSize',
];

export interface CaretCoordinates {
  top: number;
  left: number;
  height: number;
}

let mirror: HTMLDivElement | null = null;
let marker: HTMLSpanElement | null = null;

const ensureMirror = (): { mirror: HTMLDivElement; marker: HTMLSpanElement } => {
  if (mirror && marker) return { mirror, marker };
  mirror = document.createElement('div');
  const s = mirror.style;
  s.position = 'absolute';
  s.visibility = 'hidden';
  s.top = '0';
  s.left = '-9999px';
  s.whiteSpace = 'pre-wrap';
  s.wordWrap = 'break-word';
  s.overflowWrap = 'break-word';
  marker = document.createElement('span');
  document.body.appendChild(mirror);
  return { mirror, marker };
};

export const getCaretCoordinates = (
  textarea: HTMLTextAreaElement,
  position: number,
): CaretCoordinates => {
  const { mirror: m, marker: k } = ensureMirror();
  const computed = window.getComputedStyle(textarea);
  const style = m.style as unknown as Record<string, string>;

  for (const prop of MIRRORED_PROPERTIES) {
    const value = computed[prop];
    if (typeof value === 'string') {
      style[prop as string] = value;
    }
  }

  m.textContent = textarea.value.slice(0, position);
  k.textContent = textarea.value[position] || '.';
  m.appendChild(k);

  const top = k.offsetTop;
  const left = k.offsetLeft;
  const height = parseInt(computed.lineHeight, 10) || k.offsetHeight;

  return { top, left, height };
};
