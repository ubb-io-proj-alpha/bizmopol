<template>
    <section class="view-section">
        <h1>Kursy &amp; Portal Członkowski</h1>
        <div class="grid-fill">
            <div class="feature-card"><h3>Hosting</h3><p>Struktura kursu: moduły, lekcje, media</p></div>
            <div class="feature-card"><h3>Dostęp</h3><p>Dostęp po zakupie lub podpisaniu dokumentu</p></div>
            <div class="feature-card"><h3>Egzaminy</h3><p>Quizy, certyfikaty i materiały dodatkowe</p></div>
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
