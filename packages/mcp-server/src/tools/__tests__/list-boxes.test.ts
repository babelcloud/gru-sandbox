import { describe, it, expect, vi } from 'vitest';
import { handleListBoxes } from '../index';

describe('handleListBoxes', () => {
  it('should return a list of boxes', async () => {
    const mockLog = vi.fn();
    const handler = handleListBoxes(mockLog);
    
    const result = await handler({}, {}, {
      sessionId: 'test-session',
      signal: new AbortController().signal
    });
    
    expect(result).toBeDefined();
    expect(result.content).toBeDefined();
    expect(Array.isArray(result.content)).toBe(true);
    expect(result.content[0].type).toBe('text');
  });
}); 