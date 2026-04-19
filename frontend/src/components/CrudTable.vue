<script setup>
const props = defineProps({
    columns: { type: Array, required: true },
    rows: { type: Array, required: true },
    sortBy: { type: String, default: "" },
    sortDir: { type: String, default: "asc" },
    loading: { type: Boolean, default: false },
    emptyText: { type: String, default: "Brak danych." },
})

const emit = defineEmits(["sort", "action"])

function onSort(col) {
    if (col.sortable) emit("sort", col.key)
}
</script>

<template>
    <div v-if="loading" class="loading">Ładowanie...</div>
    <div v-else-if="rows.length === 0" class="empty">{{ emptyText }}</div>
    <div v-else class="table-wrap">
        <table class="crud-table">
            <thead>
                <tr>
                    <th
                        v-for="col in columns"
                        :key="col.key"
                        :class="{ sortable: col.sortable }"
                        @click="onSort(col)"
                    >
                        {{ col.label }}
                        <span v-if="col.sortable && sortBy === col.key" class="material-icons th-sort-icon">
                            {{ sortDir === "asc" ? "arrow_upward" : "arrow_downward" }}
                        </span>
                    </th>
                    <th>Akcje</th>
                </tr>
            </thead>
            <tbody>
                <tr v-for="row in rows" :key="row.id">
                    <td
                        v-for="col in columns"
                        :key="col.key"
                        :class="col.cellClass || ''"
                    >
                        <slot :name="'cell-' + col.key" :row="row">
                            {{ row[col.key] }}
                        </slot>
                    </td>
                    <td class="actions-cell">
                        <slot name="actions" :row="row" />
                    </td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<style scoped>
.loading, .empty { color: #94a3b8; padding: 40px; text-align: center; }
.table-wrap { overflow-x: auto; border-radius: 12px; border: 1px solid #334155; }
.crud-table { width: 100%; border-collapse: collapse; }
.crud-table th {
    background: #1e293b; padding: 12px 16px; text-align: left;
    color: #94a3b8; font-size: 0.85rem; font-weight: 600; text-transform: uppercase;
    border-bottom: 1px solid #334155;
}
.crud-table th.sortable { cursor: pointer; user-select: none; }
.crud-table th.sortable:hover { color: #38bdf8; }
.th-sort-icon { font-size: 0.9rem; vertical-align: middle; margin-left: 4px; }
.crud-table td { padding: 14px 16px; border-bottom: 1px solid #1e293b; color: #e2e8f0; font-size: 0.95rem; }
.crud-table tr:last-child td { border-bottom: none; }
.crud-table tr:hover td { background: #1e293b44; }
.actions-cell { display: flex; gap: 8px; }
</style>
