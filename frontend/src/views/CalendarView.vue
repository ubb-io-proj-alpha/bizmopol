<template>
    <section class="view-section">
        <h1>Kalendarz i Spotkania</h1>
        <div class="grid-fill">
            <div class="feature-card"><h3>Booking</h3><p>Rezerwacja spotkań z poziomu lejka/landing page</p></div>
            <div class="feature-card"><h3>Sync</h3><p>Synchronizacja z Google, Outlook, Zoom i Teams</p></div>
            <div class="feature-card"><h3>Strefy</h3><p>Inteligentny wybór strefy czasowej i blokowanie godzin</p></div>
        </div>
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
.grid-fill { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; margin-top: 30px; }
.feature-card { background: #1e293b; padding: 25px; border-radius: 12px; border: 1px solid #334155; }
h1 { font-size: 2.8rem; margin-bottom: 10px; color: #f8fafc; }
h3 { color: #38bdf8; margin-bottom: 10px; }
</style>
