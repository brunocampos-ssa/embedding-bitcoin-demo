import { useState } from 'react';
import { api } from '../../api/client';
import type { DepositQuote, DepositRail, Treasury } from '../../models';
import { useI18n } from '../../i18n/context';

// TreasuryPanel shows the payout treasury's balance and lets the organizer fund
// it with a deposit before any payout can be prepared. This makes the payment
// lifecycle start where it really begins — with money arriving — instead of
// assuming a pre-funded wallet.
export function TreasuryPanel({ treasury, onChanged }: { treasury?: Treasury; onChanged: () => void }) {
  const { t } = useI18n();
  const [rail, setRail] = useState<DepositRail>('lightning');
  const [amount, setAmount] = useState('1000');
  const [quote, setQuote] = useState<DepositQuote>();
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);

  const balance = treasury?.balanceSats ?? 0;
  const amountSats = Number(amount) || 0;

  async function createDeposit() {
    setBusy(true);
    setError('');
    setCopied(false);
    try {
      setQuote(await api.deposit(rail, amountSats));
      onChanged();
      // The mock credits a simulated deposit after ~1s; refresh the balance so
      // the organizer watches it arrive.
      setTimeout(onChanged, 1200);
    } catch (e) {
      setError(e instanceof Error ? e.message : t('error'));
    } finally {
      setBusy(false);
    }
  }

  async function copy() {
    if (!quote) return;
    try {
      await navigator.clipboard.writeText(quote.paymentRequest);
      setCopied(true);
    } catch {
      /* clipboard unavailable; the request is still shown for manual copy */
    }
  }

  return (
    <section className="treasury">
      <div className="treasuryHead">
        <div>
          <small>{t('treasuryBalance')}</small>
          <p className="balance">{balance.toLocaleString()} sats</p>
        </div>
        {balance === 0 && <p className="fundCue">{t('fundToBegin')}</p>}
      </div>
      <details className="depositBox" open={balance === 0}>
        <summary>{t('depositTitle')}</summary>
        <div className="depositForm">
          <label className="field">
            {t('depositRail')}
            <select value={rail} onChange={(e) => setRail(e.target.value as DepositRail)}>
              <option value="lightning">{t('railLightning')}</option>
              <option value="bitcoin">{t('railBitcoin')}</option>
              <option value="spark">{t('railSpark')}</option>
            </select>
          </label>
          <label className="field">
            {t('depositAmount')}
            <input type="number" min="1" value={amount} onChange={(e) => setAmount(e.target.value)} />
          </label>
          <button onClick={createDeposit} disabled={busy || amountSats <= 0}>
            {t('createDeposit')}
          </button>
        </div>
        {quote && (
          <div className="depositQuote">
            <small>{t('depositRequest')}</small>
            <code>{quote.paymentRequest}</code>
            <button className="textButton" onClick={copy}>
              {copied ? t('copied') : t('copy')}
            </button>
            <small>{t('depositHint')}</small>
          </div>
        )}
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
      </details>
    </section>
  );
}
