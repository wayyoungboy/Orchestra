/**
 * Strip ANSI / OSC / common orphan CSI tokens so PTY output is readable as plain chat text.
 */
const orphanTTY = /\[(?:\?[0-9]+[A-Za-z]+|[0-9][\d;]*[A-Za-z])/g

export function stripAnsiForChat(input: unknown): string {
  if (typeof input !== 'string') return ''
  if (!input) return ''
  let s = input.replace(/\r/g, '')
  let prev = ''
  while (s !== prev) {
    prev = s
    s = s.replace(/\u001b\[(?:[\x30-\x3F]*[\x20-\x2F]*)?[\x40-\x7E]/g, '')
    s = s.replace(/\u009b(?:[\x30-\x3F]*[\x20-\x2F]*)?[\x40-\x7E]/g, '')
    s = s.replace(/\u001b\][^\u0007]*(?:\u0007|\u001b\\)/g, '')
  }
  s = s.replace(/\u001b[\x20-\x2F][\x30-\x7E]/g, '')
  s = s.replace(/\u001b[\x30-\x7F]/g, '')
  s = s.replace(/\[\?2026[hl]/gi, '')
  prev = ''
  while (s !== prev) {
    prev = s
    s = s.replace(orphanTTY, '')
  }
  return s.replace(/\n{3,}/g, '\n\n').trim()
}
