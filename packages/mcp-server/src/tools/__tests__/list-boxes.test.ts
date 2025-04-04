import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';

import { handleListBoxes } from '../index';

const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('handleListBoxes', () => {
  beforeEach(() => {
    mockFetch.mockResolvedValue({
      ok: true,
      json: async () => ({ 
        boxes: [
          { id: 'box1', image: 'test-image', status: 'running' }
        ] 
      }),
      text: async () => '',
      status: 200,
      statusText: 'OK'
    });
  });

  afterEach(() => {
    mockFetch.mockReset();
  });

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
    
    expect(mockFetch).toHaveBeenCalled();
  });
});  