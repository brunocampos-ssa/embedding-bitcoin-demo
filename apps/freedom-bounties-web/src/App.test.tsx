import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, it, expect, vi } from 'vitest';
import App from './App';
import { I18nProvider } from './i18n/context';
import { api } from './api/client';
import type { Payout } from './models';

vi.mock('./api/client', () => ({
  api: {
    health: vi.fn(),
    bounties: vi.fn(),
    payments: vi.fn(),
    treasury: vi.fn(),
    bounty: vi.fn(),
    deposit: vi.fn(),
  },
}));

describe('App', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(api.health).mockResolvedValue({ status: 'ok', paymentProvider: 'mock', demoMode: true, authentication: 'demo' });
    vi.mocked(api.bounties).mockResolvedValue([]);
    vi.mocked(api.treasury).mockResolvedValue({ balanceSats: 0, identity: {} });
  });

  it('renders the payment history without crashing when the API returns null', async () => {
    // Regression: a null /api/payments payload used to crash the history view
    // (history.length / history.map on null → white screen).
    vi.mocked(api.payments).mockResolvedValue(null as unknown as Payout[]);
    render(
      <I18nProvider>
        <App />
      </I18nProvider>,
    );
    await waitFor(() => expect(api.payments).toHaveBeenCalled());
    fireEvent.click(screen.getByRole('button', { name: /payment history/i }));
    expect(await screen.findByText(/no payments yet/i)).toBeInTheDocument();
  });
});
