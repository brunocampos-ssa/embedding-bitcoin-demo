import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, it, expect, vi } from 'vitest';
import { TreasuryPanel } from './TreasuryPanel';
import { I18nProvider } from '../../i18n/context';
import { api } from '../../api/client';

vi.mock('../../api/client', () => ({ api: { deposit: vi.fn() } }));

describe('TreasuryPanel', () => {
  beforeEach(() => vi.clearAllMocks());

  it('shows the balance and a fund cue when empty', () => {
    render(
      <I18nProvider>
        <TreasuryPanel treasury={{ balanceSats: 0, identity: {} }} onChanged={() => {}} />
      </I18nProvider>,
    );
    expect(screen.getByText('0 sats')).toBeInTheDocument();
    expect(screen.getByText(/fund the treasury to begin/i)).toBeInTheDocument();
  });

  it('creates a deposit and shows the payment request', async () => {
    vi.mocked(api.deposit).mockResolvedValue({ rail: 'lightning', paymentRequest: 'lnbcdepositdemo', feeSats: 0 });
    render(
      <I18nProvider>
        <TreasuryPanel treasury={{ balanceSats: 0, identity: {} }} onChanged={() => {}} />
      </I18nProvider>,
    );
    fireEvent.change(screen.getByLabelText(/amount/i), { target: { value: '500' } });
    fireEvent.click(screen.getByRole('button', { name: /create deposit/i }));
    await waitFor(() => expect(screen.getByText('lnbcdepositdemo')).toBeInTheDocument());
    expect(api.deposit).toHaveBeenCalledWith('lightning', 500);
  });
});
