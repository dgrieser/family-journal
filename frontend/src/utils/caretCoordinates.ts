// Measure the caret position inside a <textarea> by rendering a hidden mirror
// <div> with the same typography and reading the offset of a marker span placed
// at the caret index. Returns coordinates relative to the textarea's border box.

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

export const getCaretCoordinates = (
  textarea: HTMLTextAreaElement,
  position: number,
): CaretCoordinates => {
  const mirror = document.createElement('div');
  const style = mirror.style;
  const computed = window.getComputedStyle(textarea);

  style.position = 'absolute';
  style.visibility = 'hidden';
  style.top = '0';
  style.left = '-9999px';
  style.whiteSpace = 'pre-wrap';
  style.wordWrap = 'break-word';
  style.overflowWrap = 'break-word';

  for (const prop of MIRRORED_PROPERTIES) {
    const value = computed[prop];
    if (typeof value === 'string') {
      (style as unknown as Record<string, string>)[prop as string] = value;
    }
  }

  mirror.textContent = textarea.value.slice(0, position);

  const marker = document.createElement('span');
  marker.textContent = textarea.value.slice(position) || '.';
  mirror.appendChild(marker);

  document.body.appendChild(mirror);
  const top = marker.offsetTop;
  const left = marker.offsetLeft;
  const height = parseInt(computed.lineHeight, 10) || marker.offsetHeight;
  document.body.removeChild(mirror);

  return { top, left, height };
};
