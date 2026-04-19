<template>
    <section class="view-section">
        <h1>Komunikacja wielokanałowa</h1>
        <div class="list-fill">
            <div class="item">
                <span class="material-icons item-icon">email</span> Dwukierunkowa komunikacja e-mail (Gmail/Outlook Sync)
            </div>
            <div class="item">
                <span class="material-icons item-icon">smartphone</span> Automatyczne sekwencje e-mail i SMS
            </div>
            <div class="item">
                <span class="material-icons item-icon">inbox</span> Unified Inbox: Messenger, Instagram, Google Business
            </div>
        </div>
    </section>
</template>

<style scoped>
.view-section { width: 100%; max-width: 1400px; margin: 0 auto; }
.list-fill { margin-top: 30px; }
.item {
    background: #1e293b;
    padding: 20px;
    border-radius: 8px;
    margin-bottom: 15px;
    border-left: 4px solid #38bdf8;
    display: flex;
    align-items: center;
}
.item-icon { font-size: 1.3rem; margin-right: 12px; color: #38bdf8; flex-shrink: 0; }
h1 { font-size: 2.8rem; margin-bottom: 10px; color: #f8fafc; }
</style>
