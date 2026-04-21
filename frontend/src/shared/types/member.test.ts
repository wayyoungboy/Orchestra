import { describe, it, expect } from 'vitest'
import { displayMemberStatus, Member, MemberStatus } from './member'

function make(overrides: Partial<Member> = {}): Member {
  return {
    id: '1',
    workspaceId: 'w1',
    name: 'test',
    roleType: 'assistant',
    status: 'online',
    createdAt: '2024-01-01',
    ...overrides,
  }
}

describe('displayMemberStatus', () => {
  it('returns online when manual is online and connected', () => {
    const m = make({ status: 'online', connection: 'connected' })
    expect(displayMemberStatus(m)).toBe('online')
  })

  it('returns working when connection is working', () => {
    const m = make({ connection: 'working' })
    expect(displayMemberStatus(m)).toBe('working')
  })

  it('returns offline when connection is disconnected', () => {
    const m = make({ connection: 'disconnected' })
    expect(displayMemberStatus(m)).toBe('offline')
  })

  it('manual dnd overrides working connection', () => {
    const m = make({ manualStatus: 'dnd', connection: 'working' })
    expect(displayMemberStatus(m)).toBe('dnd')
  })

  it('manual offline overrides everything', () => {
    const m = make({ manualStatus: 'offline', connection: 'connected' })
    expect(displayMemberStatus(m)).toBe('offline')
  })

  it('manual working overrides connected', () => {
    const m = make({ manualStatus: 'working', connection: 'connected' })
    expect(displayMemberStatus(m)).toBe('working')
  })

  it('connecting shows as online', () => {
    const m = make({ connection: 'connecting' })
    expect(displayMemberStatus(m)).toBe('online')
  })

  it('defaults to online when no status or connection set', () => {
    const m = make({ status: undefined as unknown as MemberStatus, connection: undefined })
    expect(displayMemberStatus(m)).toBe('online')
  })
})
