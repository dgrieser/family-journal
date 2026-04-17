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

interface CachedStyles {
  width: number;
  height: number;
  lineHeight: number;
}

// WeakMap so cached textareas can be garbage-collected after their component
// unmounts. WeakRef tracks which textarea the shared mirror is currently
// styled for without pinning it.
const styleCache = new WeakMap<HTMLTextAreaElement, CachedStyles>();
let mirrorTargetRef: WeakRef<HTMLTextAreaElement> | null = null;

export const getCaretCoordinates = (
  textarea: HTMLTextAreaElement,
  position: number,
): CaretCoordinates => {
  const { mirror: m, marker: k } = ensureMirror();
  const width = textarea.offsetWidth;
  const height = textarea.offsetHeight;
  let cached = styleCache.get(textarea);
  const mirrorMatches = mirrorTargetRef?.deref() === textarea;

  if (!cached || !mirrorMatches || cached.width !== width || cached.height !== height) {
    const computed = window.getComputedStyle(textarea);
    const style = m.style as unknown as Record<string, string>;
    for (const prop of MIRRORED_PROPERTIES) {
      const value = computed[prop];
      if (typeof value === 'string') {
        style[prop as string] = value;
      }
    }
    cached = {
      width,
      height,
      lineHeight: parseFloat(computed.lineHeight),
    };
    styleCache.set(textarea, cached);
    mirrorTargetRef = new WeakRef(textarea);
  }

  m.textContent = textarea.value.slice(0, position);
  k.textContent = textarea.value[position] || '.';
  m.appendChild(k);

  return {
    top: k.offsetTop,
    left: k.offsetLeft,
    height: cached.lineHeight || k.offsetHeight,
  };
};
