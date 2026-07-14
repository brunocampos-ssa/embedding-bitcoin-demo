import { useEffect, useState } from 'react';
import { api } from './api/client';
import type { Bounty, Health, Payout, Treasury } from './models';
import { Header } from './components/Header';
import { PayoutFlow } from './features/payouts/PayoutFlow';
import { TreasuryPanel } from './features/treasury/TreasuryPanel';
import { useI18n } from './i18n/context';

export default function App() {
  const { t } = useI18n();
  const [health, setHealth] = useState<Health>();
  const [bounties, setBounties] = useState<Bounty[]>([]);
  const [selected, setSelected] = useState<Bounty>();
  const [history, setHistory] = useState<Payout[]>([]);
  const [treasury, setTreasury] = useState<Treasury>();
  const [view, setView] = useState<'bounties' | 'history'>('bounties');
  const [error, setError] = useState('');

  async function load() {
    try {
      const [h, b, p, tr] = await Promise.all([api.health(), api.bounties(), api.payments(), api.treasury()]);
      setHealth(h);
      setBounties(b);
      setHistory(p);
      setTreasury(tr);
      if (selected) setSelected(await api.bounty(selected.id));
    } catch (e) {
      setError(e instanceof Error ? e.message : t('error'));
    }
  }
  useEffect(() => {
    void load();
  }, []);

  async function choose(id: string) {
    try {
      setSelected(await api.bounty(id));
    } catch (e) {
      setError(e instanceof Error ? e.message : t('error'));
    }
  }

  return (
    <div>
      <Header demo={health?.demoMode ?? true} />
      <main>
        <section className="hero">
          <p className="eyebrow">BITCOIN PAYMENT INFRASTRUCTURE</p>
          <h1>{t('tagline')}</h1>
          <p>{t('auth')}</p>
        </section>
        <TreasuryPanel treasury={treasury} onChanged={load} />
        <nav>
          <button className={view === 'bounties' ? 'active' : ''} onClick={() => setView('bounties')}>
            {t('bounties')}
          </button>
          <button className={view === 'history' ? 'active' : ''} onClick={() => setView('history')}>
            {t('history')}
          </button>
        </nav>
        {error && (
          <p role="alert" className="error">
            {error}
          </p>
        )}
        {view === 'history' ? (
          <section className="cards">
            {history.length === 0 ? (
              <p>{t('empty')}</p>
            ) : (
              history.map((p) => (
                <article key={p.id}>
                  <span className="pill">{p.state}</span>
                  <h3>{p.amountBaseUnits} sats</h3>
                  <p>
                    {p.rail} · {p.destinationMasked}
                  </p>
                </article>
              ))
            )}
          </section>
        ) : selected ? (
          <>
            <button className="textButton" onClick={() => setSelected(undefined)}>
              ← {t('back')}
            </button>
            <article className="detail">
              <span className="pill">{selected.state}</span>
              <h2>{selected.title}</h2>
              <p>{selected.description}</p>
              <div className="facts">
                <span>
                  {selected.rewardSats} sats · {t('reward')}
                </span>
                <span>{selected.format}</span>
                <span>{selected.language}</span>
              </div>
              {selected.submissions?.[0] && (
                <PayoutFlow
                  submission={selected.submissions[0]}
                  balanceSats={treasury?.balanceSats ?? 0}
                  rewardSats={selected.rewardSats}
                  onChanged={load}
                />
              )}
            </article>
          </>
        ) : (
          <section className="cards">
            {bounties.map((b) => (
              <article key={b.id}>
                <span className="pill">{b.state}</span>
                <h2>{b.title}</h2>
                <p>{b.description}</p>
                <strong>{b.rewardSats} sats</strong>
                <button onClick={() => choose(b.id)}>{t('open')}</button>
              </article>
            ))}
          </section>
        )}
        <aside>{t('stablecoins')}</aside>
      </main>
    </div>
  );
}
