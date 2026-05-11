<script lang="ts">
    import { ws } from "$lib/ws.svelte";
    import SensorChart from "$lib/components/SensorChart.svelte";

    const temp = $derived(ws.latest?.temperature ?? null);
    const hum = $derived(ws.latest?.humidity ?? null);
    const ts = $derived(
        ws.latest
            ? new Date(ws.latest.timestamp).toLocaleTimeString("vi-VN", {
                  hour12: false,
              })
            : "--:--:--",
    );

    const todayDate = new Date().toLocaleDateString("vi-VN", {
        weekday: "long",
        year: "numeric",
        month: "long",
        day: "numeric",
    });
</script>

<main>
    <header>
        <div class="brand">
            <svg
                width="22"
                height="22"
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg">
                <path
                    d="M12 2C8.13 2 5 5.13 5 9c0 2.38 1.19 4.47 3 5.74V17c0 .55.45 1 1 1h6c.55 0 1-.45 1-1v-2.26C17.81 13.47 19 11.38 19 9c0-3.87-3.13-7-7-7z"
                    fill="#f97316" />
                <path
                    d="M9 21c0 .55.45 1 1 1h4c.55 0 1-.45 1-1v-1H9v1z"
                    fill="#8891aa" />
            </svg>
            <span>TempLog</span>
        </div>

        <div class="status-group">
            <div class="date-label">{todayDate}</div>
            <div class="status" class:online={ws.connected}>
                <span class="dot"></span>
                {ws.connected ? "Live" : "Connecting..."}
            </div>
        </div>
    </header>

    <section class="stat-row">
        <div class="stat-card temp">
            <div class="stat-label">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                    <path
                        d="M12 2a3 3 0 0 1 3 3v8.17A5 5 0 1 1 7 17V5a3 3 0 0 1 3-3h2z"
                        stroke="#f97316"
                        stroke-width="2"
                        stroke-linecap="round" />
                </svg>
                Temperature
            </div>
            <div class="stat-value">
                {temp !== null ? temp.toFixed(2) : "---"}
                <span class="unit">°C</span>
            </div>
            <div class="stat-range">
                <div class="range-item">
                    <span class="range-label">MIN</span>
                    <span class="range-val">{ws.stats.min_temp.toFixed(1)}°</span>
                </div>
                <div class="range-item">
                    <span class="range-label">MAX</span>
                    <span class="range-val">{ws.stats.max_temp.toFixed(1)}°</span>
                </div>
            </div>
        </div>

        <div class="stat-card hum">
            <div class="stat-label">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none">
                    <path
                        d="M12 2l-6.5 9a6.5 6.5 0 1 0 13 0L12 2z"
                        stroke="#38bdf8"
                        stroke-width="2"
                        stroke-linecap="round"
                        stroke-linejoin="round" />
                </svg>
                Humidity
            </div>
            <div class="stat-value">
                {hum !== null ? hum.toFixed(2) : "---"}
                <span class="unit">%</span>
            </div>
            <div class="stat-range">
                <div class="range-item">
                    <span class="range-label">MIN</span>
                    <span class="range-val">{ws.stats.min_hum.toFixed(1)}%</span>
                </div>
                <div class="range-item">
                    <span class="range-label">MAX</span>
                    <span class="range-val">{ws.stats.max_hum.toFixed(1)}%</span>
                </div>
            </div>
        </div>

        <div class="stat-card meta">
            <div class="stat-label">Last updated</div>
            <div class="stat-ts">{ts}</div>
            <div class="stat-label" style="margin-top: 12px">Data points</div>
            <div class="stat-count">
                {ws.history.length}
                <span class="unit">/ 60</span>
            </div>
        </div>
    </section>

    <section class="chart-section">
        <div class="chart-card">
            <div class="chart-header">
                <span class="chart-title temp-text">Temperature</span>
                <span class="chart-badge temp-badge">°C</span>
            </div>
            <div class="chart-wrap">
                <SensorChart
                    history={ws.history}
                    field="temperature"
                    color="#f97316"
                    glow="rgba(249,115,22,0.08)"
                    unit="°C" />
            </div>
        </div>

        <div class="chart-card">
            <div class="chart-header">
                <span class="chart-title hum-text">Humidity</span>
                <span class="chart-badge hum-badge">%RH</span>
            </div>
            <div class="chart-wrap">
                <SensorChart
                    history={ws.history}
                    field="humidity"
                    color="#38bdf8"
                    glow="rgba(56,189,248,0.08)"
                    unit="%" />
            </div>
        </div>
    </section>
</main>

<style>
    main {
        max-width: 1200px;
        margin: 0 auto;
        padding: 24px 20px 48px;
        display: flex;
        flex-direction: column;
        gap: 24px;
    }

    /* ── header ── */
    header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .brand {
        display: flex;
        align-items: center;
        gap: 10px;
        font-size: 1.1rem;
        font-weight: 600;
        letter-spacing: -0.02em;
        color: var(--text-primary);
    }

    .status-group {
        display: flex;
        align-items: center;
        gap: 16px;
    }

    .date-label {
        font-size: 0.9rem;
        font-weight: 500;
        color: var(--text-secondary);
    }

    .status {
        display: flex;
        align-items: center;
        gap: 7px;
        font-size: 0.8rem;
        font-weight: 500;
        color: var(--offline);
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.2);
        padding: 5px 12px;
        border-radius: 999px;
        transition: all 0.3s;
    }

    .status.online {
        color: var(--online);
        background: rgba(34, 197, 94, 0.1);
        border-color: rgba(34, 197, 94, 0.2);
    }

    .dot {
        width: 7px;
        height: 7px;
        border-radius: 50%;
        background: currentColor;
    }

    .status.online .dot {
        animation: pulse 2s infinite;
    }

    @keyframes pulse {
        0%,
        100% {
            opacity: 1;
            transform: scale(1);
        }
        50% {
            opacity: 0.5;
            transform: scale(0.8);
        }
    }

    /* ── stat cards ── */
    .stat-row {
        display: grid;
        grid-template-columns: 1fr 1fr 220px;
        gap: 16px;
    }

    .stat-card {
        background: var(--bg-surface);
        border: 1px solid var(--border);
        border-radius: var(--radius-md);
        padding: 20px 22px;
        display: flex;
        flex-direction: column;
        gap: 10px;
        transition:
            border-color 0.2s,
            box-shadow 0.2s;
    }

    .stat-card.temp:hover {
        border-color: rgba(249, 115, 22, 0.4);
        box-shadow: 0 0 20px rgba(249, 115, 22, 0.07);
    }
    .stat-card.hum:hover {
        border-color: rgba(56, 189, 248, 0.4);
        box-shadow: 0 0 20px rgba(56, 189, 248, 0.07);
    }

    .stat-label {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 0.75rem;
        font-weight: 500;
        color: var(--text-secondary);
        text-transform: uppercase;
        letter-spacing: 0.06em;
    }

    .stat-value {
        font-size: 2.6rem;
        font-weight: 700;
        font-family: var(--font-mono);
        letter-spacing: -0.03em;
        line-height: 1;
        color: var(--text-primary);
    }

    .stat-card.temp .stat-value {
        color: #f97316;
    }
    .stat-card.hum .stat-value {
        color: #38bdf8;
    }

    .unit {
        font-size: 0.45em;
        font-weight: 400;
        color: var(--text-secondary);
        vertical-align: super;
        margin-left: 2px;
    }

    .stat-range {
        display: flex;
        gap: 16px;
        margin-bottom: 4px;
    }

    .range-item {
        display: flex;
        flex-direction: column;
        gap: 2px;
    }

    .range-label {
        font-size: 0.6rem;
        font-weight: 600;
        color: var(--text-secondary);
        letter-spacing: 0.05em;
    }

    .range-val {
        font-size: 0.95rem;
        font-weight: 600;
        font-family: var(--font-mono);
        color: var(--text-primary);
    }

    .stat-ts {
        font-family: var(--font-mono);
        font-size: 1.3rem;
        font-weight: 500;
        color: var(--text-primary);
    }

    .stat-count {
        font-family: var(--font-mono);
        font-size: 1.3rem;
        font-weight: 500;
        color: var(--text-primary);
    }

    /* ── charts ── */
    .chart-section {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 16px;
    }

    .chart-card {
        background: var(--bg-surface);
        border: 1px solid var(--border);
        border-radius: var(--radius-md);
        padding: 20px 22px;
        display: flex;
        flex-direction: column;
        gap: 16px;
    }

    .chart-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }

    .chart-title {
        font-size: 0.85rem;
        font-weight: 600;
        letter-spacing: 0.01em;
    }

    .temp-text {
        color: #f97316;
    }
    .hum-text {
        color: #38bdf8;
    }

    .chart-badge {
        font-family: var(--font-mono);
        font-size: 0.72rem;
        padding: 2px 8px;
        border-radius: 4px;
        font-weight: 500;
    }

    .temp-badge {
        background: rgba(249, 115, 22, 0.12);
        color: #f97316;
    }
    .hum-badge {
        background: rgba(56, 189, 248, 0.12);
        color: #38bdf8;
    }

    .chart-wrap {
        height: 220px;
        position: relative;
    }

    /* ── responsive ── */
    @media (max-width: 900px) {
        .stat-row {
            grid-template-columns: 1fr 1fr;
        }
        .stat-card.meta {
            grid-column: 1 / -1;
            flex-direction: row;
            gap: 24px;
        }
        .chart-section {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 540px) {
        .stat-row {
            grid-template-columns: 1fr;
        }
        .stat-value {
            font-size: 2rem;
        }
    }
</style>
