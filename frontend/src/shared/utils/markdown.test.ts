/** @vitest-environment jsdom */
import { describe, expect, it } from 'vitest'
import { renderMarkdownSafe } from './markdown'

describe('renderMarkdownSafe', () => {
  it('renders Markdown but treats raw HTML as text', () => {
    const rendered = renderMarkdownSafe('**safe** <img src=x onerror=alert(1)>')

    expect(rendered).toContain('<strong>safe</strong>')
    expect(rendered).not.toContain('<img')
  })

  it('removes unsafe link destinations', () => {
    expect(renderMarkdownSafe('[unsafe](javascript:alert(1))')).not.toContain('href=')
    expect(renderMarkdownSafe('[safe](https://example.com)')).toContain('href="https://example.com"')
  })
})
