import { marked } from 'marked'

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

    // Render Markdown
    const rendered = marked.parse(text, { async: false }) as string
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
 * Sanitize HTML to prevent XSS attacks
 * Removes script tags and dangerous attributes
 */
export function sanitizeHtml(html: string): string {
  // Remove script tags
  let sanitized = html.replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '')

  // Remove on* event handlers
  sanitized = sanitized.replace(/\s*on\w+\s*=\s*["'][^"']*["']/gi, '')

  // Remove javascript: URLs
  sanitized = sanitized.replace(/javascript:/gi, '')

  return sanitized
}

/**
 * Render Markdown with sanitization for safe display
 */
export function renderMarkdownSafe(text: string): string {
  const rendered = renderMarkdown(text)
  return sanitizeHtml(rendered)
}