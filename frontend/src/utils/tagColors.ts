/**
 * Golden angle color assignment for persons and hashtags.
 * hue_i = (i * 137.508) % 360
 *
 * The djb2 hash converts a word into a stable integer index,
 * then the golden angle step ensures maximum perceptual separation.
 */

/** Shared regex for splitting / matching @mention and #hashtag tokens. */
export const TAG_PATTERN = /([@#][\p{L}\d_]+)/gu;

export function goldenAngleHue(word: string): number {
  let hash = 5381;
  const s = word.toLowerCase();
  for (let i = 0; i < s.length; i++) {
    hash = ((hash << 5) + hash) ^ s.charCodeAt(i);
    hash = hash >>> 0; // keep unsigned 32-bit
  }
  return (hash * 137.508) % 360;
}

export interface TagColors {
  color: string;
  background: string;
  border: string;
}

export function getTagColors(word: string): TagColors {
  const hue = goldenAngleHue(word);
  return {
    color: `hsl(${hue}, 65%, 35%)`,
    background: `hsl(${hue}, 80%, 92%)`,
    border: `hsl(${hue}, 55%, 80%)`,
  };
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

/**
 * Converts raw post text into an HTML string where @mentions and #hashtags
 * are wrapped in <mark> tags with golden-angle inline styles.
 * Used by the PostForm textarea backdrop overlay.
 * All non-tag text is HTML-escaped to prevent XSS.
 */
export function buildHighlightHtml(text: string): string {
  const parts = text.split(TAG_PATTERN);
  return parts
    .map((part) => {
      if (/^[@#][\p{L}\d_]+$/u.test(part)) {
        const name = part.slice(1);
        const hue = goldenAngleHue(name);
        // Use box-shadow instead of padding/border so the mark adds zero
        // horizontal width — keeping text metrics identical to the textarea
        // and preserving cursor position alignment.
        return (
          `<mark style="` +
          `background:hsl(${hue},80%,92%);` +
          `color:hsl(${hue},65%,35%);` +
          `border-radius:3px;` +
          `box-shadow:0 0 0 3px hsl(${hue},80%,92%),0 0 0 4px hsl(${hue},55%,80%)` +
          `">${escapeHtml(part)}</mark>`
        );
      }
      return escapeHtml(part);
    })
    .join('');
}
