/**
 * Stable hue assignment for persons and hashtags.
 *
 * Runs FNV-1a and djb2 in one pass, then combines them with a Fibonacci
 * mix step. Two independent hashes break the clustering that any single hash
 * produces for short, structurally similar words (e.g. German tags sharing
 * common character ranges).
 */

/** Shared regex for splitting / matching @mention and #hashtag tokens. */
export const TAG_PATTERN = /([@#][\p{L}\d_]+)/gu;

export function goldenAngleHue(word: string): number {
  const s = word.toLowerCase();
  let h1 = 2166136261; // FNV-1a 32-bit offset basis
  let h2 = 5381;       // djb2 seed
  for (let i = 0; i < s.length; i++) {
    const c = s.charCodeAt(i);
    // FNV-1a step
    h1 ^= c;
    h1 = Math.imul(h1, 16777619); // FNV-1a prime; Math.imul avoids float precision loss
    h1 = h1 >>> 0;
    // djb2 step (independent second hash)
    h2 = Math.imul(h2, 33) ^ c;
    h2 = h2 >>> 0;
  }
  // Combine: multiply h2 by the 32-bit Fibonacci/golden-ratio constant before XOR
  // so that correlated low bits in h1 and h2 land in different output positions.
  return ((h1 ^ Math.imul(h2, 0x9e3779b9)) >>> 0) % 360;
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
          `box-shadow:0 0 0 1px hsl(${hue},55%,80%)` +
          `">${escapeHtml(part)}</mark>`
        );
      }
      return escapeHtml(part);
    })
    .join('');
}
