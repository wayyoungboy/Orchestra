import { marked } from 'marked'

const ALLOWED_TAGS = new Set([
  'a', 'blockquote', 'br', 'code', 'del', 'em', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6',
  'hr', 'li', 'ol', 'p', 'pre', 'strong', 'table', 'tbody', 'td', 'th', 'thead', 'tr', 'ul',
])

const ALLOWED_ATTRIBUTES: Record<string, Set<string>> = {
  a: new Set(['href', 'title']),
  code: new Set(['class']),
  ol: new Set(['start']),
}

// Configure marked options for safe rendering
marked.setOptions({
  breaks: true,        // Convert line breaks to <br>
  gfm: true,           // GitHub Flavored Markdown
})

/**
 * Render Markdown text to HTML with fallback
 * @param text - Raw text (may contain Markdown)
 * @returns Rendered HTML string, or original text if rendering fails
 */
export function renderMarkdown(text: string): string {
  if (!text) return ''

  try {
    // Check if text contains any Markdown indicators
    const hasMarkdown = containsMarkdown(text)

    if (!hasMarkdown) {
      // No Markdown detected, return as-is (with line break handling)
      return escapeHtml(text).replace(/\n/g, '<br>')
    }

    // Raw HTML is not part of the chat markdown dialect. Escaping it before
    // parsing prevents marked from turning attacker-controlled tags into DOM.
    const rendered = marked.parse(escapeHtml(text), { async: false }) as string
    return rendered
  } catch (error) {
    // Fallback: return escaped plain text
    console.warn('Markdown rendering failed, using plain text:', error)
    return escapeHtml(text).replace(/\n/g, '<br>')
  }
}

/**
 * Check if text likely contains Markdown syntax
 */
function containsMarkdown(text: string): boolean {
  // Common Markdown patterns
  const markdownPatterns = [
    /^#{1,6}\s/m,           // Headers
    /\*\*.*?\*\*/,          // Bold
    /\*.*?\*/,              // Italic
    /`[^`]+`/,              // Inline code
    /```[\s\S]*?```/,       // Code blocks
    /^\s*[-*+]\s/m,         // Unordered lists
    /^\s*\d+\.\s/m,         // Ordered lists
    /\[.*?\]\(.*?\)/,       // Links
    /^\s*>/m,               // Blockquotes
    /---+/,                 // Horizontal rules
    /\|.+\|/,               // Tables
  ]

  return markdownPatterns.some(pattern => pattern.test(text))
}

/**
 * Escape HTML special characters to prevent XSS
 */
function escapeHtml(text: string): string {
  const htmlEntities: Record<string, string> = {
    '&': '&amp;',
    '<': '&lt;',
    '>': '&gt;',
    '"': '&quot;',
    "'": '&#39;',
  }

  return text.replace(/[&<>"']/g, char => htmlEntities[char] || char)
}

/**
 * Sanitize marked's output with a strict element and attribute allowlist.
 * This is deliberately a whitelist: removing a few dangerous strings is not
 * enough to make arbitrary HTML safe.
 */
export function sanitizeHtml(html: string): string {
  if (typeof document === 'undefined') return escapeHtml(html)

  const template = document.createElement('template')
  template.innerHTML = html

  const sanitizeNode = (node: Node) => {
    if (!(node instanceof Element)) return

    const tag = node.tagName.toLowerCase()
    if (!ALLOWED_TAGS.has(tag)) {
      node.replaceWith(document.createTextNode(node.textContent || ''))
      return
    }

    const allowedAttributes = ALLOWED_ATTRIBUTES[tag] || new Set<string>()
    for (const attribute of Array.from(node.attributes)) {
      if (!allowedAttributes.has(attribute.name)) {
        node.removeAttribute(attribute.name)
        continue
      }
      if (attribute.name === 'href' && !isSafeLink(attribute.value)) {
        node.removeAttribute(attribute.name)
      }
      if (tag === 'code' && attribute.name === 'class' && !/^language-[a-z0-9+-]+$/i.test(attribute.value)) {
        node.removeAttribute(attribute.name)
      }
    }

    for (const child of Array.from(node.childNodes)) sanitizeNode(child)
  }

  for (const child of Array.from(template.content.childNodes)) sanitizeNode(child)
  return template.innerHTML
}

function isSafeLink(href: string): boolean {
  try {
    const url = new URL(href, window.location.origin)
    return url.protocol === 'http:' || url.protocol === 'https:' || url.protocol === 'mailto:'
  } catch {
    return false
  }
}

/**
 * Render Markdown with sanitization for safe display
 */
export function renderMarkdownSafe(text: string): string {
  const rendered = renderMarkdown(text)
  return sanitizeHtml(rendered)
}
